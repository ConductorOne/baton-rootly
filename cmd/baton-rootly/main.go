package main

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/conductorone/baton-rootly/pkg/config"
	"github.com/conductorone/baton-rootly/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-rootly",
		getConnector,
		cfg.Config,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// getConnector initializes and returns the connector.
func getConnector(ctx context.Context, rc *cfg.Rootly) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	apiKey := rc.ApiKey

	c, err := connector.New(ctx, apiKey)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	server, err := connectorbuilder.NewConnector(ctx, c)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return server, nil
}
