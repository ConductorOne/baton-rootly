package connector

import (
	"context"
	"time"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Users.List",
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

	// fetch users from the Rootly API with pagination
	users, token, err := o.client.GetUsers(ctx, bag.PageToken())
	if err != nil {
		return nil, "", nil, err
	}

	// create user resources using the SDK
	var resources []*v2.Resource
	for _, user := range users {
		userResource, err := sdkResource.NewUserResource(
			getBestName(user.Attributes),
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

	// set the next page token
	nextPage, err := bag.NextToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// getUserTraitOptions returns a list of UserTraitOption's based on the available fields for a Rootly user.
func getUserTraitOptions(user client.User) []sdkResource.UserTraitOption {
	traitOpts := []sdkResource.UserTraitOption{
		sdkResource.WithEmail(user.Attributes.Email, true),
		// always set status to enabled, since Rootly doesn't allow for disabled user status
		sdkResource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
		sdkResource.WithUserProfile(getUserProfile(user)),
	}
	if t, err := time.Parse(time.RFC3339, user.Attributes.CreatedAt); err == nil {
		traitOpts = append(traitOpts, sdkResource.WithCreatedAt(t))
	}
	return traitOpts
}

// getUserProfile builds a map of profile fields from the available user fields.
func getUserProfile(user client.User) map[string]interface{} {
	// required Rootly fields
	profile := map[string]interface{}{
		"user_id":    user.ID,
		"updated_at": user.Attributes.UpdatedAt,
	}

	// optional Rootly fields
	if user.Attributes.Name != "" {
		profile["name"] = user.Attributes.Name
	}
	if user.Attributes.FullName != "" {
		profile["full_name"] = user.Attributes.FullName
		first, last := sdkResource.SplitFullName(user.Attributes.FullName)
		profile["first_name"] = first
		profile["last_name"] = last
	}
	if user.Attributes.SlackID != "" {
		profile["slack_id"] = user.Attributes.SlackID
	}
	if user.Attributes.Phone != "" {
		profile["phone"] = user.Attributes.Phone
	}
	return profile
}

// getBestName checks the user fields for the best name-like field that is also populated.
// Defaults to email as a fallback, which is a required field in Rootly.
func getBestName(userAttr client.UserAttributes) string {
	if userAttr.Name != "" {
		return userAttr.Name
	}
	if userAttr.FullName != "" {
		return userAttr.FullName
	}
	return userAttr.Email
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
	}
}
