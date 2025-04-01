package connector

import (
	"context"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	// TODO: Parse pagination token if needed

	// fetch users from the Rootly API with pagination
	users, nextPage, err := o.client.GetUsers(ctx, pToken)
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, user := range users {
		// create the user resource using the SDK helper
		userResource, err := sdkResource.NewUserResource(
			getBestName(user),
			userResourceType,
			user.ID,
			getUserTraitOptions(user),
			sdkResource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, userResource)
	}

	return resources, nextPage, nil, nil
}

// getUserTraitOptions returns a list of UserTraitOption based on the available fields for a Rootly user.
func getUserTraitOptions(user client.User) []sdkResource.UserTraitOption {
	// required Rootly fields
	profile := map[string]interface{}{
		// TODO: confirm correct keys to use
		"user_id": user.ID,
	}
	// optional Rootly fields
	if user.Name != "" {
		profile["login"] = user.Name // should key be "name"?
	}
	if user.FullName != "" {
		profile["full_name"] = user.FullName
		first, last := sdkResource.SplitFullName(user.FullName)
		profile["first_name"] = first
		profile["last_name"] = last
	}
	if user.SlackID != "" {
		profile["slack_id"] = user.SlackID
	}
	if user.Phone != "" {
		profile["phone"] = user.Phone
	}
	return []sdkResource.UserTraitOption{
		sdkResource.WithEmail(user.Email, true),
		// always set status to enabled, since Rootly doesn't allow for disabled user status
		sdkResource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
		sdkResource.WithCreatedAt(user.CreatedAt),
		// should we add user.UpdatedAt ? (always present)
		sdkResource.WithUserProfile(profile),
	}
}

// getBestName checks the user fields for the best name-like field that is also populated.
// Defaults to email as a fallback, which is a required field in Rootly.
func getBestName(user client.User) string {
	if user.Name != "" {
		return user.Name
	}
	if user.FullName != "" {
		return user.FullName
	}
	return user.Email
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{
		client: client,
	}
}
