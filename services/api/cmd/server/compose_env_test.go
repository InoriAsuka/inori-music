package main

// Guards against two related, opposite-direction classes of drift between
// docker-compose.prod.yml's "api" service environment and the
// INORI_*/MEILI_* variables cmd/server/main.go actually reads via
// os.Getenv:
//
//   - TestComposeAPIEnvVarsMatchSource (the bug fixed in v5.34.0): compose
//     declares a name the source never reads. docker-compose.prod.yml used
//     to declare INORI_DB_DSN and INORI_AUTH_SECRET, neither of which
//     cmd/server/main.go ever reads (it reads INORI_DATABASE_URL, and no
//     source file reads an auth secret by that name at all). The container
//     started, the health check passed, and the API silently ran with
//     in-memory, non-persistent user/session storage — a deploy-time
//     footgun that produced no error anywhere.
//
//   - TestComposeAPIEnvVarsIncludeSilentlyDegradingVars (the bug fixed in
//     v5.34.1): the mirror image — source reads a name whose absence
//     silently swaps in a materially worse fallback, but compose never
//     declared it at all. docker-compose.prod.yml ran a whole "meilisearch"
//     service that the api service was never wired up to use:
//     MEILI_HOST/MEILI_SEARCH_KEY were absent from the api service's
//     environment, so main.go's `if meiliHost := os.Getenv("MEILI_HOST");
//     meiliHost != ""` check was always false, catalog search silently fell
//     back to PostgreSQL full-text search, and — unlike most of this
//     binary's other config-dependent fallbacks — not even a log line
//     marked the fact.
//
// Both are purely static checks: they parse Go source and YAML, and never
// start a container or open a database connection.

import (
	"fmt"
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

// trackedEnvPrefixes are the environment-variable name prefixes the tests in
// this file reason about — the two prefixes cmd/server/main.go actually
// reads via os.Getenv (confirmed by grepping both exhaustively during the
// v5.34.0/v5.34.1 compose-vs-source audits).
var trackedEnvPrefixes = []string{"INORI_", "MEILI_"}

func hasTrackedPrefix(name string) bool {
	for _, p := range trackedEnvPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

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

// sourceEnvVars returns every INORI_/MEILI_-prefixed name read via
// os.Getenv(...) literal calls in the given source file.
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
		if hasTrackedPrefix(name) {
			vars[name] = true
		}
		return true
	})
	return vars
}

// composeAPIEnvVarNames returns the INORI_/MEILI_-prefixed variable names
// declared in docker-compose.prod.yml's "api" service environment mapping.
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
		if hasTrackedPrefix(key) {
			names = append(names, key)
		}
	}
	sort.Strings(names)
	return names
}

// TestComposeAPIEnvVarsMatchSource cross-checks docker-compose.prod.yml's
// "api" service environment against cmd/server/main.go — the entry point
// that service actually runs (see the repo-root Dockerfile: it builds
// ./services/api/cmd/server). Every INORI_*/MEILI_* name declared in
// compose must be a name the binary actually reads; otherwise it is dead or
// misspelled and will silently no-op in production instead of failing
// loudly.
//
// This intentionally checks only one direction (compose -> source). The
// reverse would fail the build every time an optional env var (e.g. E2E
// test fixtures) legitimately has no entry in a production compose file —
// that direction is instead covered, for a curated subset of names where
// "legitimately has no entry" is not actually legitimate, by
// TestComposeAPIEnvVarsIncludeSilentlyDegradingVars below.
func TestComposeAPIEnvVarsMatchSource(t *testing.T) {
	repoRoot := findRepoRoot(t)
	sourceVars := sourceEnvVars(t, filepath.Join(repoRoot, "services", "api", "cmd", "server", "main.go"))
	if len(sourceVars) == 0 {
		t.Fatal("found zero INORI_*/MEILI_* os.Getenv(...) calls in cmd/server/main.go — this test's parser is broken, not the source")
	}

	composeVars := composeAPIEnvVarNames(t, filepath.Join(repoRoot, "docker-compose.prod.yml"))
	if len(composeVars) == 0 {
		t.Fatal("found zero INORI_*/MEILI_* keys in docker-compose.prod.yml's api service environment — this test's YAML parsing is broken, not the compose file")
	}

	var unknown []string
	for _, name := range composeVars {
		if !sourceVars[name] {
			unknown = append(unknown, name)
		}
	}

	if len(unknown) > 0 {
		t.Fatalf("docker-compose.prod.yml's api service declares INORI_*/MEILI_* variables cmd/server/main.go never reads via os.Getenv — these are dead or misspelled and will silently no-op in production: %s",
			strings.Join(unknown, ", "))
	}
}

// silentlyDegradingEnvVars is a hand-maintained registry of environment
// variables cmd/server/main.go reads where an *unset* value does not fail
// loudly or even necessarily log clearly — it silently swaps in a
// materially worse runtime behavior. Every name here must appear as a key
// in docker-compose.prod.yml's api service environment section.
//
// This is intentionally a curated list, not something inferred by scanning
// source for every os.Getenv call (TestComposeAPIEnvVarsMatchSource already
// covers the opposite direction). Most INORI_* variables this binary reads
// are legitimately fine left unset in production, with a fully-functional
// documented default or fallback and no capability actually lost — e.g.
// INORI_CORS_ORIGINS (permissive CORS, by design) and INORI_SESSION_TTL
// (documented 24h default). Requiring every var source reads to appear in
// compose would false-positive on those and get "fixed" by deleting the
// test rather than the config. Registering a name here is a judgment call
// about whether unset causes a real loss of function, which static
// analysis cannot make on its own — see v5.34.1's meilisearch wiring bug
// for the shape of problem this exists to catch: MEILI_HOST/MEILI_SEARCH_KEY
// were absent from compose, catalog search silently fell back from
// Meilisearch to PostgreSQL full-text search, and the meilisearch container
// this same compose file started kept running, permanently unused, with no
// log line ever pointing at why.
//
// When you add a new "unset means quietly do something materially worse"
// env var to main.go, register it here with a one-line note on what unset
// silently does. If it is a legitimate optional toggle with a sane,
// fully-functional default instead, it does not belong in this map.
var silentlyDegradingEnvVars = map[string]string{
	"MEILI_HOST": "catalog search silently falls back from Meilisearch to " +
		"PostgreSQL full-text search with no log line at all — main.go's " +
		"MEILI_HOST != \"\" check is simply false, so the whole meilisearch " +
		"init block is skipped — and the meilisearch service in this same " +
		"compose file keeps running, permanently unused.",
	"MEILI_SEARCH_KEY": "meilisearch requests authenticate with an empty " +
		"API key; since this compose file's meilisearch service always " +
		"requires MEILI_MASTER_KEY, Meilisearch responds unauthorized, " +
		"NewMeilisearch's health check fails, and search falls back to " +
		"PostgreSQL full-text search — the same practical outcome as " +
		"MEILI_HOST being unset, just reached via a failed health check " +
		"(which does log) instead of a silent early exit.",
}

// TestComposeAPIEnvVarsIncludeSilentlyDegradingVars asserts every variable
// registered in silentlyDegradingEnvVars is wired into docker-compose.prod
// .yml's api service environment. See that map's doc comment for what
// qualifies and why this is a curated allowlist rather than an automatic
// reverse of TestComposeAPIEnvVarsMatchSource.
//
// It also self-checks the registry against main.go: a registered name that
// main.go no longer reads via os.Getenv is a stale entry (e.g. after a
// rename) and fails loudly rather than silently checking for the wrong
// thing forever.
func TestComposeAPIEnvVarsIncludeSilentlyDegradingVars(t *testing.T) {
	repoRoot := findRepoRoot(t)
	sourceVars := sourceEnvVars(t, filepath.Join(repoRoot, "services", "api", "cmd", "server", "main.go"))

	composeVars := composeAPIEnvVarNames(t, filepath.Join(repoRoot, "docker-compose.prod.yml"))
	inCompose := make(map[string]bool, len(composeVars))
	for _, v := range composeVars {
		inCompose[v] = true
	}

	names := make([]string, 0, len(silentlyDegradingEnvVars))
	for name := range silentlyDegradingEnvVars {
		names = append(names, name)
	}
	sort.Strings(names)

	var stale, missing []string
	for _, name := range names {
		if !sourceVars[name] {
			stale = append(stale, name)
		}
		if !inCompose[name] {
			missing = append(missing, name)
		}
	}

	if len(stale) > 0 {
		t.Fatalf("silentlyDegradingEnvVars in this file registers %s, but cmd/server/main.go no longer reads it via os.Getenv(...) — fix the stale registry entry (rename or remove it)",
			strings.Join(stale, ", "))
	}
	if len(missing) > 0 {
		var detail strings.Builder
		for _, name := range missing {
			fmt.Fprintf(&detail, "\n  %s — unset, %s", name, silentlyDegradingEnvVars[name])
		}
		t.Fatalf("docker-compose.prod.yml's api service environment is missing variables that silently degrade functionality when left unset (registered in silentlyDegradingEnvVars in this file):%s",
			detail.String())
	}
}
