package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-rootly/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-rootly",
		getConnector,
		field.Configuration{
			Fields: ConfigurationFields,
		},
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

// getConnector initializes and returns the connector
func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)
	if err := ValidateConfig(v); err != nil {
		return nil, err
	}

	apiKey := v.GetString(APIKeyField.FieldName)

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
