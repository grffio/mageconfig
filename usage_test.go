package mageconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHelpRequested(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "no args",
			args:     []string{"prog"},
			expected: false,
		},
		{
			name:     "unrelated args",
			args:     []string{"prog", "arg1", "arg2"},
			expected: false,
		},
		{
			name:     "short help flag",
			args:     []string{"prog", "-help"},
			expected: true,
		},
		{
			name:     "long help flag",
			args:     []string{"prog", "--help"},
			expected: true,
		},
		{
			name:     "mixed flags",
			args:     []string{"prog", "arg1", "--help", "arg2"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Beware: this modifies global state and is not safe for parallel test execution.
			os.Args = tc.args
			got := isHelpRequested()
			assert.Equal(t, tc.expected, got)
		})
	}
}
