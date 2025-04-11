package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	scheduleOwnerEntitlement  = "owner"
	scheduleMemberEntitlement = "member"
	scheduleOnCallEntitlement = "on-call"
)

type scheduleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *scheduleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the schedules from the database as resource objects.
// Schedules include a GroupTrait because they are the 'shape' of a standard group.
func (o *scheduleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Schedules.List",
		zap.String("pToken", pToken.Token),
	)

	// set up pagination
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}
	// initialize pagination state if needed
	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: o.resourceType.Id,
		})
	}

	// fetch schedules from the Rootly API with pagination
	schedules, token, err := o.client.GetSchedules(ctx, bag.PageToken())
	if err != nil {
		return nil, "", nil, err
	}

	// create schedule resources using the SDK
	var resources []*v2.Resource
	for _, schedule := range schedules {
		scheduleResource, err := sdkResource.NewGroupResource(
			schedule.Attributes.Name,
			o.resourceType,
			schedule.ID,
			getScheduleTraitOptions(schedule),
			sdkResource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, scheduleResource)
	}

	// set the next page token
	nextPage, err := bag.NextToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// getScheduleTraitOptions returns a list of GroupTraitOption's based on the available fields for a Rootly schedule.
func getScheduleTraitOptions(schedule client.Schedule) []sdkResource.GroupTraitOption {
	// required Rootly fields
	profile := map[string]interface{}{
		"schedule_id": schedule.ID,
		"name":        schedule.Attributes.Name,
		"created_at":  schedule.Attributes.CreatedAt,
		"updated_at":  schedule.Attributes.UpdatedAt,
	}

	// optional Rootly fields
	if schedule.Attributes.Description != "" {
		profile["description"] = schedule.Attributes.Description
	}

	return []sdkResource.GroupTraitOption{
		sdkResource.WithGroupProfile(profile),
	}
}

// Entitlements for each schedule include ownership, membership, and on-call membership.
func (o *scheduleBuilder) Entitlements(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Schedules.Entitlements",
		zap.String("resource.DisplayName", resource.DisplayName),
		zap.String("resource.Id.Resource", resource.Id.Resource),
	)

	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			scheduleOwnerEntitlement,
			entitlement.WithGrantableTo(userResourceType, teamResourceType),
			entitlement.WithDisplayName(
				fmt.Sprintf("%s schedule owner", resource.DisplayName)),
			entitlement.WithDescription(
				fmt.Sprintf("Is owner of the %s schedule in Rootly", resource.DisplayName),
			),
		),
		entitlement.NewAssignmentEntitlement(
			resource,
			scheduleMemberEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(
				fmt.Sprintf("%s schedule member", resource.DisplayName),
			),
			entitlement.WithDescription(
				fmt.Sprintf("Is member of the %s schedule in Rootly", resource.DisplayName),
			),
		),
		entitlement.NewAssignmentEntitlement(
			resource,
			scheduleOnCallEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(
				fmt.Sprintf("%s schedule on-call member", resource.DisplayName),
			),
			entitlement.WithDescription(
				fmt.Sprintf("Is on-call member of the %s schedule in Rootly", resource.DisplayName),
			),
		),
	}, "", nil, nil
}

// Grants for each schedule include checking current owners, current members, and current on-call members.
func (o *scheduleBuilder) Grants(
	ctx context.Context,
	resource *v2.Resource,
	pToken *pagination.Token,
) ([]*v2.Grant, string, annotations.Annotations, error) {
	// set up pagination
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}
	// initialize pagination state if needed
	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: resource.Id.ResourceType,
			ResourceID:     resource.Id.Resource,
		})
	}
	// create a new resource type ID for tracking schedule rotation pagination, only used within Grants
	const scheduleRotationResourceTypeID = "schedule_rotation"

	var grants []*v2.Grant
	switch bag.ResourceTypeID() {
	case scheduleResourceType.Id:
		scheduleID := bag.ResourceID()
		// only handle the schedule owners and on-call members once, ie on the first iteration
		if bag.PageToken() == "" {
			// fetch schedule owners from the Rootly API
			ownerUserID, ownerTeamIDs, err := o.client.GetScheduleOwnerIDs(ctx, scheduleID)
			if err != nil {
				return nil, "", nil, err
			}
			// add a grant for the owner user
			if ownerUserID != nil {
				grants = append(grants, grant.NewGrant(
					resource,
					scheduleOwnerEntitlement,
					&v2.ResourceId{
						ResourceType: userResourceType.Id,
						Resource:     strconv.Itoa(*ownerUserID),
					},
				))
			}
			// add grants for the owner team(s), and the users nested within
			for _, ownerTeamID := range ownerTeamIDs {
				grants = append(grants, grant.NewGrant(
					resource,
					scheduleOwnerEntitlement,
					&v2.ResourceId{
						ResourceType: teamResourceType.Id,
						Resource:     ownerTeamID,
					},
					grant.WithAnnotation(&v2.GrantExpandable{
						EntitlementIds: []string{
							fmt.Sprintf("team:%s:%s", ownerTeamID, teamMemberEntitlement),
							fmt.Sprintf("team:%s:%s", ownerTeamID, teamAdminEntitlement),
						},
					}),
				))
			}

			// fetch schedule on-call members from the Rootly API
			onCallUserIDs, err := o.client.ListOnCallUsers(ctx, scheduleID)
			if err != nil {
				return nil, "", nil, err
			}
			// add grants for schedule on-call members
			for _, onCallUserID := range onCallUserIDs {
				grants = append(grants, grant.NewGrant(
					resource,
					scheduleOnCallEntitlement,
					&v2.ResourceId{
						ResourceType: userResourceType.Id,
						Resource:     strconv.Itoa(onCallUserID),
					},
				))
			}
		}

		// fetching schedule members is more complex since it entails nested paginated API calls:
		// 	1) this iteration fetch schedule rotations from the Rootly API and push each rotation to the bag.
		// 	   if there are more rotation pages, also push the next page token to the bag for a future iteration.
		// 	2) next iteration(s) fetch all the members for a rotation, handled within the other switch case.
		rotationIDs, nextPage, err := o.client.ListScheduleRotations(ctx, scheduleID, bag.PageToken())
		if err != nil {
			return nil, "", nil, err
		}
		bag.Pop()
		if nextPage != "" {
			// there are more schedule rotations to fetch for this schedule
			bag.Push(pagination.PageState{
				ResourceTypeID: scheduleResourceType.Id,
				ResourceID:     scheduleID,
				Token:          nextPage,
			})
		}
		for _, rotationID := range rotationIDs {
			bag.Push(pagination.PageState{
				ResourceTypeID: scheduleRotationResourceTypeID,
				ResourceID:     rotationID,
			})
		}
	case scheduleRotationResourceTypeID:
		// fetch all members for the schedule rotation from the Rootly API
		rotationID := bag.ResourceID()
		memberUserIDs, err := o.client.ListAllScheduleRotationUsers(ctx, rotationID)
		if err != nil {
			return nil, "", nil, err
		}
		bag.Pop()
		// add grants for these members
		for _, memberUserID := range memberUserIDs {
			grants = append(grants, grant.NewGrant(
				resource,
				scheduleMemberEntitlement,
				&v2.ResourceId{
					ResourceType: userResourceType.Id,
					Resource:     strconv.Itoa(memberUserID),
				},
			))
		}
	}

	pageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return grants, pageToken, nil, nil
}

func newScheduleBuilder(client *client.Client) *scheduleBuilder {
	return &scheduleBuilder{
		client:       client,
		resourceType: scheduleResourceType,
	}
}
