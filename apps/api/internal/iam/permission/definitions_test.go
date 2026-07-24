package permission

import "testing"

func TestAllContainsUniquePermissions(
	t *testing.T,
) {
	t.Parallel()

	seen := make(
		map[Permission]struct{},
	)

	for _, currentPermission := range All() {
		if currentPermission == "" {
			t.Fatal(
				"permission must not be empty",
			)
		}

		if _, exists := seen[currentPermission]; exists {
			t.Fatalf(
				"duplicate permission %q",
				currentPermission,
			)
		}

		seen[currentPermission] = struct{}{}
	}
}
