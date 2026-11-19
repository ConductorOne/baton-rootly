package connector

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var _ connectorbuilder.ResourceSyncer = (*roleBuilder)(nil)

// roleAssignedEntitlement is the single entitlement both role kinds expose. Rootly assigns
// exactly one Incident Response role and one On-Call role per user, so "assigned" is the
// only meaningful permission on a role.
const roleAssignedEntitlement = "assigned"

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *roleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the Incident Response roles from Rootly as resource objects.
// Roles include a RoleTrait because they are the 'shape' of a standard role.
func (o *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Roles.List",
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

	// fetch roles from the Rootly API with pagination
	roles, token, err := o.client.GetRoles(ctx, bag.PageToken())
	if err != nil {
		return nil, "", nil, err
	}

	// create role resources using the SDK
	var resources []*v2.Resource
	for _, role := range roles {
		roleResource, err := sdkResource.NewRoleResource(
			roleDisplayName(role.Attributes.Name, role.Attributes.Slug),
			o.resourceType,
			role.ID,
			nil,
			sdkResource.WithParentResourceID(parentResourceID),
			sdkResource.WithResourceProfile(getRoleProfile(role)),
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

// getRoleProfile builds a map of profile fields from the available Rootly role fields.
func getRoleProfile(role client.Role) map[string]interface{} {
	// required Rootly fields
	profile := map[string]interface{}{
		"role_id":    role.ID,
		"name":       role.Attributes.Name,
		"created_at": role.Attributes.CreatedAt,
		"updated_at": role.Attributes.UpdatedAt,
	}

	// optional Rootly fields
	if role.Attributes.Slug != "" {
		profile["slug"] = role.Attributes.Slug
	}

	return profile
}

// Entitlements for each role is the single "assigned" assignment entitlement.
func (o *roleBuilder) Entitlements(
	ctx context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Roles.Entitlements",
		zap.String("resource.DisplayName", resource.DisplayName),
		zap.String("resource.Id.Resource", resource.Id.Resource),
	)

	return []*v2.Entitlement{
		newRoleAssignedEntitlement(resource, "Incident Response role"),
	}, "", nil, nil
}

// Grants always returns an empty slice for roles. The grants live on the user side: Rootly
// returns each user's role relationship inline on /v1/users, so the user syncer emits every
// role grant in a single pass instead of re-scanning the user list once per role.
func (o *roleBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newRoleBuilder(client *client.Client) *roleBuilder {
	return &roleBuilder{
		client:       client,
		resourceType: roleResourceType,
	}
}

// newRoleAssignedEntitlement builds the "assigned" entitlement shared by both role kinds.
func newRoleAssignedEntitlement(resource *v2.Resource, roleKind string) *v2.Entitlement {
	return entitlement.NewAssignmentEntitlement(
		resource,
		roleAssignedEntitlement,
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDisplayName(fmt.Sprintf("%s Assigned", resource.DisplayName)),
		entitlement.WithDescription(fmt.Sprintf("Has the %s %s in Rootly", resource.DisplayName, roleKind)),
	)
}

// roleDisplayName qualifies a role name with its slug when the slug carries information the
// name does not. Rootly creates a system role per team, so a catalog commonly holds several
// roles all named "Admin" that are only told apart by their slug (admin, admin-1, ...).
func roleDisplayName(name string, slug string) string {
	if name == "" {
		return slug
	}
	if slug == "" || slug == slugify(name) {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, slug)
}

// slugify approximates Rootly's slug generation so roleDisplayName can tell an uninformative
// slug (the plain kebab-case of the name) from one that disambiguates duplicates.
func slugify(name string) string {
	var sb strings.Builder
	lastWasDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			sb.WriteRune(r)
			lastWasDash = false
		case !lastWasDash:
			sb.WriteRune('-')
			lastWasDash = true
		}
	}
	return strings.Trim(sb.String(), "-")
}
