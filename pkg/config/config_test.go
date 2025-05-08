package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Rootly
		wantErr bool
	}{
		{
			name: "valid config - with api key",
			config: &Rootly{
				ApiKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty api key",
			config: &Rootly{
				ApiKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), "required field 'api-key' is missing")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
