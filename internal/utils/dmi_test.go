package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsLaptop(t *testing.T) {
	tests := []struct {
		name           string
		dmiContent     string
		expectedLaptop bool
		description    string
	}{
		{
			name:           "laptop code 8 - Portable",
			dmiContent:     "8",
			expectedLaptop: true,
			description:    "Numeric chassis type 8 (Portable)",
		},
		{
			name:           "laptop code 9 - Laptop",
			dmiContent:     "9",
			expectedLaptop: true,
			description:    "Numeric chassis type 9 (Laptop)",
		},
		{
			name:           "laptop code 10 - Notebook",
			dmiContent:     "10",
			expectedLaptop: true,
			description:    "Numeric chassis type 10 (Notebook)",
		},
		{
			name:           "laptop code 14 - Sub Notebook",
			dmiContent:     "14",
			expectedLaptop: true,
			description:    "Numeric chassis type 14 (Sub Notebook)",
		},
		{
			name:           "laptop code 31 - Convertible",
			dmiContent:     "31",
			expectedLaptop: true,
			description:    "Numeric chassis type 31 (Convertible)",
		},
		{
			name:           "laptop code 32 - Detachable",
			dmiContent:     "32",
			expectedLaptop: true,
			description:    "Numeric chassis type 32 (Detachable)",
		},
		{
			name:           "desktop code 3",
			dmiContent:     "3",
			expectedLaptop: false,
			description:    "Numeric chassis type 3 (Desktop)",
		},
		{
			name:           "server code 17",
			dmiContent:     "17",
			expectedLaptop: false,
			description:    "Numeric chassis type 17 (Server)",
		},
		{
			name:           "text portable",
			dmiContent:     "Portable",
			expectedLaptop: true,
			description:    "Textual chassis type 'Portable'",
		},
		{
			name:           "text laptop lowercase",
			dmiContent:     "laptop",
			expectedLaptop: true,
			description:    "Textual chassis type 'laptop' (lowercase)",
		},
		{
			name:           "text laptop mixed case",
			dmiContent:     "Laptop",
			expectedLaptop: true,
			description:    "Textual chassis type 'Laptop' (mixed case)",
		},
		{
			name:           "text notebook lowercase",
			dmiContent:     "notebook",
			expectedLaptop: true,
			description:    "Textual chassis type 'notebook' (lowercase)",
		},
		{
			name:           "text notebook mixed case",
			dmiContent:     "Notebook",
			expectedLaptop: true,
			description:    "Textual chassis type 'Notebook' (mixed case)",
		},
		{
			name:           "text sub-notebook with hyphen",
			dmiContent:     "sub-notebook",
			expectedLaptop: true,
			description:    "Textual chassis type 'sub-notebook' (with hyphen)",
		},
		{
			name:           "text sub notebook without hyphen",
			dmiContent:     "sub notebook",
			expectedLaptop: true,
			description:    "Textual chassis type 'sub notebook' (without hyphen)",
		},
		{
			name:           "text desktop",
			dmiContent:     "Desktop",
			expectedLaptop: false,
			description:    "Textual chassis type 'Desktop'",
		},
		{
			name:           "text server",
			dmiContent:     "Server",
			expectedLaptop: false,
			description:    "Textual chassis type 'Server'",
		},
		{
			name:           "code with whitespace",
			dmiContent:     "  10  \n",
			expectedLaptop: true,
			description:    "Numeric code with leading/trailing whitespace",
		},
		{
			name:           "text with whitespace",
			dmiContent:     "  laptop  \n",
			expectedLaptop: true,
			description:    "Textual value with leading/trailing whitespace",
		},
		{
			name:           "unknown numeric code",
			dmiContent:     "99",
			expectedLaptop: false,
			description:    "Unknown numeric chassis type",
		},
		{
			name:           "unknown text value",
			dmiContent:     "Unknown",
			expectedLaptop: false,
			description:    "Unknown textual chassis type",
		},
		{
			name:           "empty file",
			dmiContent:     "",
			expectedLaptop: false,
			description:    "Empty DMI file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dmiFile := filepath.Join(tmpDir, "chassis_type")

			err := os.WriteFile(dmiFile, []byte(tt.dmiContent), 0o600)
			require.NoError(t, err, "Failed to write test DMI file")

			t.Setenv("HYPRDYNAMICMONITORS_DMI_PATH_OVERRIDE", dmiFile)

			result := utils.IsLaptop()
			assert.Equal(t, tt.expectedLaptop, result, tt.description)
		})
	}
}

func TestIsLaptop_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does_not_exist")

	t.Setenv("HYPRDYNAMICMONITORS_DMI_PATH_OVERRIDE", nonExistentFile)

	result := utils.IsLaptop()
	assert.True(t, result, "IsLaptop should return true when DMI file doesn't exist (backwards compatibility)")
}

func TestIsLaptop_NoPermission(t *testing.T) {
	tmpDir := t.TempDir()
	dmiFile := filepath.Join(tmpDir, "chassis_type")

	err := os.WriteFile(dmiFile, []byte("10"), 0o600)
	require.NoError(t, err, "Failed to write test DMI file")

	err = os.Chmod(dmiFile, 0o000)
	require.NoError(t, err, "Failed to change file permissions")

	t.Cleanup(func() {
		_ = os.Chmod(dmiFile, 0o600)
	})

	t.Setenv("HYPRDYNAMICMONITORS_DMI_PATH_OVERRIDE", dmiFile)

	result := utils.IsLaptop()
	assert.True(t, result, "IsLaptop should return true when DMI file can't be read (backwards compatibility)")
}

func TestIsLaptop_DefaultPath(t *testing.T) {
	os.Unsetenv("HYPRDYNAMICMONITORS_DMI_PATH_OVERRIDE")
	// Just verify that it doesn't panic
	result := utils.IsLaptop()
	t.Logf("IsLaptop on this system returned: %v", result)
}
