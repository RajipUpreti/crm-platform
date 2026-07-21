package auth

import "testing"

func TestNormalizeAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "host and port",
			value:    "localhost:8081",
			expected: "localhost:8081",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := normalizeAddress(test.value)
			if err != nil {
				t.Fatalf("normalizeAddress() error = %v", err)
			}

			if actual != test.expected {
				t.Fatalf(
					"normalizeAddress() = %q; expected %q",
					actual,
					test.expected,
				)
			}
		})
	}
}

func TestNormalizeAddressRejectsInvalidDialTarget(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"http://keycloak",
		"http://keycloak:8080",
		"ftp://keycloak:21",
		"http://keycloak:8080/realms/crm",
	} {
		value := value

		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if _, err := normalizeAddress(value); err == nil {
				t.Fatalf(
					"normalizeAddress(%q) accepted an invalid dial target",
					value,
				)
			}
		})
	}
}
