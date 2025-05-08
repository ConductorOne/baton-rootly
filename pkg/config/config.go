package config

import (
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIKeyField = field.StringField(
		"api-key",
		field.WithDescription("The API key for authenticating with Rootly"),
		field.WithRequired(true),
	)

	//go:generate go run ./gen
	Config = field.NewConfiguration([]field.SchemaField{
		APIKeyField,
	})

	// FieldRelationships defines relationships between the fields listed in
	// Config that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(cfg *Rootly) error {
	apiKey := cfg.GetString(APIKeyField.FieldName)
	if len(apiKey) == 0 {
		return fmt.Errorf("required field 'api-key' is missing")
	}

	return nil
}
