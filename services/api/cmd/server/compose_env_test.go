package main

// Guards against the class of bug fixed in v5.34.0: docker-compose.prod.yml
// declared INORI_DB_DSN and INORI_AUTH_SECRET, neither of which
// cmd/server/main.go ever reads (it reads INORI_DATABASE_URL, and no source
// file reads an auth secret by that name at all). The container started,
// the health check passed, and the API silently ran with in-memory,
// non-persistent user/session storage — a deploy-time footgun that produces
// no error anywhere. This test fails whenever docker-compose.prod.yml's
// "api" service declares an INORI_* variable name the deployed binary does
// not actually read via os.Getenv, catching the same class of drift before
// it reaches a real deployment.
//
// This is a purely static check: it parses Go source and YAML, and never
// starts a container or opens a database connection.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// findRepoRoot walks up from the current working directory (a Go test's
// working directory is always its package directory) looking for go.work,
// the monorepo root marker. docker-compose.prod.yml lives at the repo root,
// outside the services/api Go module, so it cannot be reached with a fixed
// "../.." relative count that would silently break if this package moves.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (looked for go.work) walking up from %s", dir)
		}
		dir = parent
	}
}

// sourceEnvVars returns every INORI_-prefixed name read via os.Getenv(...)
// literal calls in the given source file.
func sourceEnvVars(t *testing.T, path string) map[string]bool {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	vars := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Getenv" {
			return true
		}
		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok || pkgIdent.Name != "os" {
			return true
		}
		if len(call.Args) != 1 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		name, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		if strings.HasPrefix(name, "INORI_") {
			vars[name] = true
		}
		return true
	})
	return vars
}

// composeAPIEnvVarNames returns the INORI_-prefixed variable names declared
// in docker-compose.prod.yml's "api" service environment mapping.
func composeAPIEnvVarNames(t *testing.T, composePath string) []string {
	t.Helper()
	raw, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read %s: %v", composePath, err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", composePath, err)
	}

	services, ok := doc["services"].(map[string]any)
	if !ok {
		t.Fatalf("%s: no top-level services mapping", composePath)
	}
	api, ok := services["api"].(map[string]any)
	if !ok {
		t.Fatalf("%s: no services.api mapping", composePath)
	}
	env, ok := api["environment"].(map[string]any)
	if !ok {
		t.Fatalf("%s: services.api.environment is missing or not a mapping (list-form `- KEY=value` environment is not handled by this check)", composePath)
	}

	var names []string
	for key := range env {
		if strings.HasPrefix(key, "INORI_") {
			names = append(names, key)
		}
	}
	sort.Strings(names)
	return names
}

// TestComposeAPIEnvVarsMatchSource cross-checks docker-compose.prod.yml's
// "api" service environment against cmd/server/main.go — the entry point
// that service actually runs (see the repo-root Dockerfile: it builds
// ./services/api/cmd/server). Every INORI_* name declared in compose must be
// a name the binary actually reads; otherwise it is dead or misspelled and
// will silently no-op in production instead of failing loudly.
//
// This intentionally checks only one direction (compose -> source). The
// reverse would fail the build every time an optional env var (e.g. E2E
// test fixtures) legitimately has no entry in a production compose file.
func TestComposeAPIEnvVarsMatchSource(t *testing.T) {
	repoRoot := findRepoRoot(t)
	sourceVars := sourceEnvVars(t, filepath.Join(repoRoot, "services", "api", "cmd", "server", "main.go"))
	if len(sourceVars) == 0 {
		t.Fatal("found zero INORI_* os.Getenv(...) calls in cmd/server/main.go — this test's parser is broken, not the source")
	}

	composeVars := composeAPIEnvVarNames(t, filepath.Join(repoRoot, "docker-compose.prod.yml"))
	if len(composeVars) == 0 {
		t.Fatal("found zero INORI_* keys in docker-compose.prod.yml's api service environment — this test's YAML parsing is broken, not the compose file")
	}

	var unknown []string
	for _, name := range composeVars {
		if !sourceVars[name] {
			unknown = append(unknown, name)
		}
	}

	if len(unknown) > 0 {
		t.Fatalf("docker-compose.prod.yml's api service declares INORI_* variables cmd/server/main.go never reads via os.Getenv — these are dead or misspelled and will silently no-op in production: %s",
			strings.Join(unknown, ", "))
	}
}
