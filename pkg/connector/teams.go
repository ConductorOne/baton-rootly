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
	teamAdminEntitlement  = "admin"
	teamMemberEntitlement = "member"
)

type teamBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *teamBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the teams from the database as resource objects.
// Teams include a GroupTrait because they are the 'shape' of a standard group.
func (o *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Teams.List",
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

	// fetch teams from the Rootly API with pagination
	teams, token, err := o.client.GetTeams(ctx, bag.PageToken())
	if err != nil {
		return nil, "", nil, err
	}

	// create team resources using the SDK
	var resources []*v2.Resource
	for _, team := range teams {
		teamResource, err := sdkResource.NewGroupResource(
			team.Attributes.Name,
			o.resourceType,
			team.ID,
			getTeamTraitOptions(team),
			sdkResource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, teamResource)
	}

	// set the next page token
	nextPage, err := bag.NextToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// getTeamTraitOptions returns a list of GroupTraitOption's based on the available fields for a Rootly team.
func getTeamTraitOptions(team client.Team) []sdkResource.GroupTraitOption {
	// required Rootly fields
	profile := map[string]interface{}{
		"team_id":    team.ID,
		"name":       team.Attributes.Name,
		"created_at": team.Attributes.CreatedAt,
		"updated_at": team.Attributes.UpdatedAt,
	}

	// optional Rootly fields
	if team.Attributes.Description != "" {
		profile["description"] = team.Attributes.Description
	}

	return []sdkResource.GroupTraitOption{
		sdkResource.WithGroupProfile(profile),
	}
}

// Entitlements for each team include administration and membership.
func (o *teamBuilder) Entitlements(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Teams.Entitlements",
		zap.String("resource.DisplayName", resource.DisplayName),
		zap.String("resource.Id.Resource", resource.Id.Resource),
	)

	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			teamAdminEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s Team Admin", resource.DisplayName)),
			entitlement.WithDescription(fmt.Sprintf("Is admin of the %s team in Rootly", resource.DisplayName)),
		),
		entitlement.NewAssignmentEntitlement(
			resource,
			teamMemberEntitlement,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s Team Member", resource.DisplayName)),
			entitlement.WithDescription(fmt.Sprintf("Is member of the %s team in Rootly", resource.DisplayName)),
		),
	}, "", nil, nil
}

// Grants for each team are the current administration and memberships.
func (o *teamBuilder) Grants(
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
			ResourceTypeID: o.resourceType.Id,
		})
	}

	// fetch team member and admin userIDs from the Rootly API
	memberIDs, adminIDs, err := o.client.GetTeamMemberAndAdminIDs(ctx, resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	var grants []*v2.Grant
	// add grants for team members
	for _, memberID := range memberIDs {
		grants = append(grants, grant.NewGrant(
			resource,
			teamMemberEntitlement,
			&v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     strconv.Itoa(memberID),
			},
		))
	}
	// add grants for team admins
	for _, adminID := range adminIDs {
		grants = append(grants, grant.NewGrant(
			resource,
			teamAdminEntitlement,
			&v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     strconv.Itoa(adminID),
			},
		))
	}

	return grants, "", nil, nil
}

func newTeamBuilder(client *client.Client) *teamBuilder {
	return &teamBuilder{
		client:       client,
		resourceType: teamResourceType,
	}
}
