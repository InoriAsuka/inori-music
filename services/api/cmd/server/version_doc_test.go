package main

// Guards documentation drift between the VERSION file and requirement.md's
// "## Current Version" field.
//
// Discovered 2026-08-12 while reviewing v5.35.0: requirement.md still
// claimed `5.21.0` while VERSION had reached 5.35.0 — the field had gone
// unmaintained since v5.22.0, drifting through fourteen releases. Nothing
// failed, because nothing ever read it: it is a human-facing claim about
// what the repository currently is, and stale human-facing claims are worse
// than absent ones, since they are still trusted.
//
// Per-release paperwork already updates VERSION and prepends a requirement.md
// history entry; this field sits between the two and was simply forgotten
// each time. A guard is the only thing that turns "forgotten" into "the build
// tells you".

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// currentVersionRe matches the "## Current Version" heading followed by the
// backtick-quoted semver on a later line, tolerating blank lines between them
// so the guard tracks the documented value rather than the exact whitespace.
var currentVersionRe = regexp.MustCompile("(?m)^##\\s+Current Version\\s*$\\s*`([^`]+)`")

func TestRequirementCurrentVersionMatchesVersionFile(t *testing.T) {
	repoRoot := findRepoRoot(t)

	rawVersion, err := os.ReadFile(filepath.Join(repoRoot, "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	version := strings.TrimSpace(string(rawVersion))
	if version == "" {
		t.Fatal("VERSION is empty")
	}

	rawDoc, err := os.ReadFile(filepath.Join(repoRoot, "requirement.md"))
	if err != nil {
		t.Fatalf("read requirement.md: %v", err)
	}

	match := currentVersionRe.FindSubmatch(rawDoc)
	if match == nil {
		t.Fatal("requirement.md has no \"## Current Version\" section with a backtick-quoted version; " +
			"if the section was renamed or removed, update this guard rather than deleting it")
	}
	documented := strings.TrimSpace(string(match[1]))

	if documented != version {
		t.Errorf("requirement.md \"## Current Version\" is %q but the VERSION file says %q — "+
			"the release paperwork updated VERSION without updating requirement.md", documented, version)
	}
}
