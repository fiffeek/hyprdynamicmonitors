package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid lowercase",
			profileName: "home",
			wantErr:     false,
		},
		{
			name:        "valid with numbers",
			profileName: "office2",
			wantErr:     false,
		},
		{
			name:        "valid with hyphens",
			profileName: "home-office",
			wantErr:     false,
		},
		{
			name:        "valid with underscores",
			profileName: "home_office",
			wantErr:     false,
		},
		{
			name:        "valid complex",
			profileName: "my-profile_123",
			wantErr:     false,
		},
		{
			name:        "valid starts with uppercase",
			profileName: "Home",
			wantErr:     false,
		},
		{
			name:        "valid all uppercase",
			profileName: "HOME",
			wantErr:     false,
		},
		{
			name:        "valid camelCase",
			profileName: "homeOffice",
			wantErr:     false,
		},
		{
			name:        "valid PascalCase",
			profileName: "HomeOffice",
			wantErr:     false,
		},
		{
			name:        "valid mixed case with numbers",
			profileName: "MyProfile123",
			wantErr:     false,
		},
		{
			name:        "empty string",
			profileName: "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "starts with number",
			profileName: "1home",
			wantErr:     true,
			errContains: "must start with a letter",
		},
		{
			name:        "contains space",
			profileName: "home office",
			wantErr:     true,
			errContains: "letters, numbers, hyphens, and underscores",
		},
		{
			name:        "contains special characters",
			profileName: "home@office",
			wantErr:     true,
			errContains: "letters, numbers, hyphens, and underscores",
		},
		{
			name:        "too long",
			profileName: "this_is_a_very_long_profile_name_that_exceeds_the_maximum_allowed_length",
			wantErr:     true,
			errContains: "must be 50 characters or less",
		},
		{
			name:        "starts with hyphen",
			profileName: "-home",
			wantErr:     true,
			errContains: "must start with a letter",
		},
		{
			name:        "starts with underscore",
			profileName: "_home",
			wantErr:     true,
			errContains: "must start with a letter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfileName(tt.profileName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
