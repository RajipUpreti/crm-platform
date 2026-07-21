package auth

import (
	"net/url"
	"testing"
)

func TestIsSafeLocalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "dashboard path",
			value:    "/dashboard",
			expected: true,
		},
		{
			name:     "nested local path",
			value:    "/app/acme/contacts?page=2",
			expected: true,
		},
		{
			name:     "absolute external URL",
			value:    "https://example.com",
			expected: false,
		},
		{
			name:     "protocol relative URL",
			value:    "//example.com/path",
			expected: false,
		},
		{
			name:     "relative path without leading slash",
			value:    "dashboard",
			expected: false,
		},
		{
			name:     "empty value",
			value:    "",
			expected: false,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := isSafeLocalPath(test.value)

			if actual != test.expected {
				t.Fatalf(
					"isSafeLocalPath(%q) = %v; expected %v",
					test.value,
					actual,
					test.expected,
				)
			}
		})
	}
}

func TestFrontendDestination(t *testing.T) {
	t.Parallel()

	frontendURL, err := url.Parse("https://crm.example.com")
	if err != nil {
		t.Fatalf("parse frontend URL: %v", err)
	}

	handler := &Handler{frontendURL: frontendURL}

	destination, err := handler.frontendDestination(
		"/app/acme/contacts?page=2",
	)
	if err != nil {
		t.Fatalf("frontendDestination() error = %v", err)
	}

	const expected = "https://crm.example.com/app/acme/contacts?page=2"
	if destination != expected {
		t.Fatalf(
			"frontendDestination() = %q; expected %q",
			destination,
			expected,
		)
	}
}

func TestFrontendDestinationRejectsExternalURL(t *testing.T) {
	t.Parallel()

	frontendURL, err := url.Parse("https://crm.example.com")
	if err != nil {
		t.Fatalf("parse frontend URL: %v", err)
	}

	handler := &Handler{frontendURL: frontendURL}

	if _, err := handler.frontendDestination(
		"https://attacker.example.com",
	); err == nil {
		t.Fatal("frontendDestination() accepted an external URL")
	}
}
