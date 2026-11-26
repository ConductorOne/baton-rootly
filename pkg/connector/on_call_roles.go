package connector

import (
	"context"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var _ connectorbuilder.ResourceSyncer = (*onCallRoleBuilder)(nil)

type onCallRoleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *onCallRoleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the On-Call roles from Rootly as resource objects.
// On-Call roles include a RoleTrait because they are the 'shape' of a standard role.
func (o *onCallRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to OnCallRoles.List",
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

	// fetch on-call roles from the Rootly API with pagination
	roles, token, err := o.client.GetOnCallRoles(ctx, bag.PageToken())
	if err != nil {
		return nil, "", nil, err
	}

	// create on-call role resources using the SDK
	var resources []*v2.Resource
	for _, role := range roles {
		roleResource, err := sdkResource.NewRoleResource(
			roleDisplayName(role.Attributes.Name, role.Attributes.Slug),
			o.resourceType,
			role.ID,
			nil,
			sdkResource.WithParentResourceID(parentResourceID),
			sdkResource.WithResourceProfile(getOnCallRoleProfile(role)),
		)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, roleResource)
	}

	// set the next page token
	nextPage, err := bag.NextToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// getOnCallRoleProfile builds a map of profile fields from the available Rootly on-call role fields.
func getOnCallRoleProfile(role client.OnCallRole) map[string]interface{} {
	// required Rootly fields
	profile := map[string]interface{}{
		"on_call_role_id": role.ID,
		"name":            role.Attributes.Name,
		"created_at":      role.Attributes.CreatedAt,
		"updated_at":      role.Attributes.UpdatedAt,
	}

	// optional Rootly fields
	if role.Attributes.Slug != "" {
		profile["slug"] = role.Attributes.Slug
	}
	// system_role is one of admin, user, custom, observer, no_access. It is the field a
	// reviewer needs to tell a stock Admin role from a bespoke custom one.
	if role.Attributes.SystemRole != "" {
		profile["system_role"] = role.Attributes.SystemRole
	}

	return profile
}

// Entitlements for each on-call role is the single "assigned" assignment entitlement.
func (o *onCallRoleBuilder) Entitlements(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to OnCallRoles.Entitlements",
		zap.String("resource.DisplayName", resource.DisplayName),
		zap.String("resource.Id.Resource", resource.Id.Resource),
	)

	return []*v2.Entitlement{
		newRoleAssignedEntitlement(resource, "On-Call role"),
	}, "", nil, nil
}

// Grants always returns an empty slice for on-call roles. As with Incident Response roles the
// grants are emitted by the user syncer, which reads the relationship inline off /v1/users.
func (o *onCallRoleBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newOnCallRoleBuilder(client *client.Client) *onCallRoleBuilder {
	return &onCallRoleBuilder{
		client:       client,
		resourceType: onCallRoleResourceType,
	}
}
