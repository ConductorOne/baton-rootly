package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-rootly/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *client.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client),
		newTeamBuilder(d.client),
		newSecretBuilder(d.client),
		newScheduleBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(_ context.Context, _ *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(_ context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Rootly Connector",
		Description: "Connector for syncing Rootly resources to Baton",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	if d.client.IsTest() {
		// skip for capabilities and config generation
		return nil, nil
	}

	_, _, err := d.client.GetUsers(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("rootly client validation failed: %w", err)
	}

	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, apiKey string) (*Connector, error) {
	rootlyClient, err := client.NewClient(ctx, client.BaseURLStr, apiKey, client.ResourcesPageSize)
	if err != nil {
		return nil, err
	}
	return &Connector{
		client: rootlyClient,
	}, nil
}
