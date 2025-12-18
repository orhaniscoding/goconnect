package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationsFS(t *testing.T) {
	entries, err := FS.ReadDir(".")
	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	foundUp := false
	foundDown := false

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if len(name) > 3 && name[len(name)-3:] == "sql" {
				if name[len(name)-7:] == ".up.sql" {
					foundUp = true
				}
				if name[len(name)-9:] == ".down.sql" {
					foundDown = true
				}
			}
		}
	}

	assert.True(t, foundUp, "Should have at least one .up.sql migration")
	// Down migrations are optional but usually present in pairs
	assert.True(t, foundDown, "Should have at least one .down.sql migration")
}
