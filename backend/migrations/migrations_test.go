package migrations

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var recentMigrationDownCompanionAllowlist = map[string]struct{}{}

func TestRecentMigrationsHaveDownCompanions(t *testing.T) {
	files, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)

	fileSet := make(map[string]struct{}, len(files))
	for _, name := range files {
		fileSet[name] = struct{}{}
	}

	for _, name := range files {
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		version := migrationNumericPrefix(name)
		if version < 1 {
			continue
		}
		if _, ok := recentMigrationDownCompanionAllowlist[name]; ok {
			continue
		}
		downName := strings.TrimSuffix(name, ".sql") + ".down.sql"
		_, ok := fileSet[downName]
		require.Truef(t, ok, "missing down companion for %s", name)
	}
}

func TestRecentSchemaMigrationsPreferIdempotentDDL(t *testing.T) {
	files, err := fs.Glob(FS, "*.sql")
	require.NoError(t, err)

	for _, name := range files {
		if migrationNumericPrefix(name) < 74 {
			continue
		}

		raw, err := fs.ReadFile(FS, name)
		require.NoErrorf(t, err, "read %s", name)
		content := strings.ToUpper(string(raw))

		assertIdempotentPattern(t, name, content, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS")
		assertOneOfPatterns(t, name, content, "CREATE INDEX", "CREATE INDEX IF NOT EXISTS", "CREATE INDEX CONCURRENTLY IF NOT EXISTS")
		assertIdempotentPattern(t, name, content, "DROP TABLE", "DROP TABLE IF EXISTS")
		assertOneOfPatterns(t, name, content, "DROP INDEX", "DROP INDEX IF EXISTS", "DROP INDEX CONCURRENTLY IF EXISTS")
		assertIdempotentPattern(t, name, content, "DROP COLUMN", "DROP COLUMN IF EXISTS")

		if strings.Contains(content, "ADD COLUMN") {
			require.Containsf(t, content, "ADD COLUMN IF NOT EXISTS", "%s should use IF NOT EXISTS for ADD COLUMN", name)
		}
	}
}

func assertIdempotentPattern(t *testing.T, name, content, keyword, expected string) {
	t.Helper()
	if strings.Contains(content, keyword) {
		require.Containsf(t, content, expected, "%s should prefer idempotent DDL: %s", name, expected)
	}
}

func assertOneOfPatterns(t *testing.T, name, content, keyword string, expected ...string) {
	t.Helper()
	if !strings.Contains(content, keyword) {
		return
	}
	for _, pattern := range expected {
		if strings.Contains(content, pattern) {
			return
		}
	}
	require.Failf(t, "missing idempotent DDL pattern", "%s should prefer one of: %s", name, strings.Join(expected, ", "))
}

func migrationNumericPrefix(name string) int {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0
	}
	value, err := strconv.Atoi(prefix)
	if err != nil {
		return 0
	}
	return value
}
