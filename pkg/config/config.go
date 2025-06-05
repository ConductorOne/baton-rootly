package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIKeyField = field.StringField(
		"api-key",
		field.WithDisplayName("API key"),
		field.WithDescription("The API key for authenticating with Rootly"),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)

	//go:generate go run ./gen
	Config = field.NewConfiguration(
		[]field.SchemaField{APIKeyField},
		field.WithConnectorDisplayName("Rootly"),
		field.WithHelpUrl("/docs/baton/rootly"),
		field.WithIconUrl("/static/app-icons/rootly.svg"),
	)

	// FieldRelationships defines relationships between the fields listed in
	// Config that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)
