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

type secretBuilder struct {
	resourceType *v2.ResourceType
	client       *client.Client
}

func (o *secretBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// List returns all the secrets from the database as resource objects.
func (o *secretBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Starting call to Secrets.List",
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

	// fetch secrets from the Rootly API with pagination
	secrets, token, err := o.client.GetSecrets(ctx, bag.PageToken())
	if err != nil {
		return nil, "", nil, err
	}

	// create secret resources using the SDK
	var resources []*v2.Resource
	for _, secret := range secrets {
		secretResource, err := sdkResource.NewSecretResource(
			secret.Attributes.Name,
			o.resourceType,
			secret.ID,
			getSecretTraitOptions(secret),
			sdkResource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, secretResource)
	}

	// set the next page token
	nextPage, err := bag.NextToken(token)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// getSecretTraitOptions returns a list of SecretTraitOption's based on the available fields for a Rootly secret.
func getSecretTraitOptions(secret client.Secret) []sdkResource.SecretTraitOption {
	var traitOpts []sdkResource.SecretTraitOption
	if t, err := time.Parse(time.RFC3339, secret.Attributes.CreatedAt); err == nil {
		traitOpts = append(traitOpts, sdkResource.WithSecretCreatedAt(t))
	}
	return traitOpts
}

// Entitlements always returns an empty slice for secrets.
func (o *secretBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for secrets since they don't have any entitlements.
func (o *secretBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newSecretBuilder(client *client.Client) *secretBuilder {
	return &secretBuilder{
		client:       client,
		resourceType: secretResourceType,
	}
}
