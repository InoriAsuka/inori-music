package main

// Guards against the class of bug fixed in v5.34.0: a new internal/*/postgres
// repository package can define an exported Migrate function that nothing
// ever calls, so its tables silently never get created against a live
// database. playerstate/postgres and searchhistory/postgres both shipped
// this way from v5.4.0 onward — NewRepository only stores the pool, it does
// not lazily create tables, and neither package had a test that would have
// exercised the real Postgres path. See userplaylistpg.Migrate in main.go
// for the pattern every new postgres package's wiring must follow.
//
// This is a purely static check: it parses source with go/parser and never
// opens a database connection.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// postgresPackageMigrate describes one internal/*/postgres package that
// declares an exported, package-level Migrate function.
type postgresPackageMigrate struct {
	importPath string // e.g. "inori-music/services/api/internal/playerstate/postgres"
	pkgName    string // e.g. "playerstatepg" (its `package` clause)
}

// findPostgresMigrateFuncs scans internal/*/postgres for packages declaring
// an exported, package-level `func Migrate(...)`. Methods (funcs with a
// receiver) are deliberately excluded: only a directly callable package
// function counts, matching how every Migrate in this codebase is written
// and called today.
func findPostgresMigrateFuncs(t *testing.T, apiRoot string) []postgresPackageMigrate {
	t.Helper()

	dirs, err := filepath.Glob(filepath.Join(apiRoot, "internal", "*", "postgres"))
	if err != nil {
		t.Fatalf("glob internal/*/postgres: %v", err)
	}
	if len(dirs) == 0 {
		t.Fatal("no internal/*/postgres directories found; glob pattern is broken, not the source tree")
	}

	fset := token.NewFileSet()
	var found []postgresPackageMigrate
	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatalf("glob %s: %v", dir, err)
		}
		pkgName := ""
		hasMigrate := false
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			node, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", f, err)
			}
			pkgName = node.Name.Name
			for _, decl := range node.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil {
					continue // methods don't count, only a directly-callable package func
				}
				if fn.Name.Name == "Migrate" {
					hasMigrate = true
				}
			}
		}
		if !hasMigrate {
			continue
		}
		rel, err := filepath.Rel(apiRoot, dir)
		if err != nil {
			t.Fatalf("rel %s: %v", dir, err)
		}
		found = append(found, postgresPackageMigrate{
			importPath: "inori-music/services/api/" + filepath.ToSlash(rel),
			pkgName:    pkgName,
		})
	}
	return found
}

// TestPostgresMigrateFuncsAreCalledInMain statically verifies that every
// internal/*/postgres package exporting a Migrate function is both imported
// and actually invoked (as "<alias>.Migrate(...)") in cmd/server/main.go —
// the only place database migrations run for the deployed API binary.
//
// This does not prove a migration runs correctly against a real database
// (that needs an actual Postgres instance); it proves the call site exists,
// which is exactly what was missing for playerstate/postgres and
// searchhistory/postgres before v5.34.0.
func TestPostgresMigrateFuncsAreCalledInMain(t *testing.T) {
	apiRoot := filepath.Join("..", "..") // cmd/server -> cmd -> services/api
	withMigrate := findPostgresMigrateFuncs(t, apiRoot)
	if len(withMigrate) == 0 {
		t.Fatal("expected at least one internal/*/postgres package exporting Migrate (e.g. storage/postgres); found none — this test's discovery logic is broken")
	}

	mainPath := filepath.Join(apiRoot, "cmd", "server", "main.go")
	mainSrc, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read %s: %v", mainPath, err)
	}
	fset := token.NewFileSet()
	mainFile, err := parser.ParseFile(fset, mainPath, mainSrc, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", mainPath, err)
	}

	// Map each imported internal/*/postgres package's import path to the
	// local identifier main.go uses for it (import alias, or the package's
	// own name when unaliased).
	aliasByImportPath := map[string]string{}
	for _, imp := range mainFile.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !strings.HasPrefix(path, "inori-music/services/api/internal/") || !strings.HasSuffix(path, "/postgres") {
			continue
		}
		if imp.Name != nil {
			aliasByImportPath[path] = imp.Name.Name
		} else {
			parts := strings.Split(path, "/")
			aliasByImportPath[path] = parts[len(parts)-1]
		}
	}

	// Collect every "<ident>.Migrate(...)" call expression in main.go.
	calledIdents := map[string]bool{}
	ast.Inspect(mainFile, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Migrate" {
			return true
		}
		if ident, ok := sel.X.(*ast.Ident); ok {
			calledIdents[ident.Name] = true
		}
		return true
	})

	var missing []string
	for _, pkg := range withMigrate {
		alias, imported := aliasByImportPath[pkg.importPath]
		switch {
		case !imported:
			missing = append(missing, pkg.importPath+" (package "+pkg.pkgName+") exports Migrate but is not imported in main.go")
		case !calledIdents[alias]:
			missing = append(missing, pkg.importPath+" (imported as \""+alias+"\") exports Migrate but main.go never calls "+alias+".Migrate(...)")
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("new PostgreSQL tables would silently never be created — wire the migration in cmd/server/main.go next to where the repository is constructed (see playerstatepg.Migrate / userplaylistpg.Migrate for the pattern):\n  %s",
			strings.Join(missing, "\n  "))
	}
}
