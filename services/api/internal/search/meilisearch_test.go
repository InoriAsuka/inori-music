package search

import (
	"encoding/json"
	"testing"
)

func TestParseSearchHitReturnsHighlightWhenPresent(t *testing.T) {
	hit := map[string]json.RawMessage{
		"id":         json.RawMessage(`"t1"`),
		"_formatted": json.RawMessage(`{"title": "<mark>World</mark> Is Mine"}`),
	}
	id, snippet := parseSearchHit(hit, "title")
	if id != "t1" {
		t.Fatalf("id = %q, want t1", id)
	}
	if snippet != "<mark>World</mark> Is Mine" {
		t.Fatalf("snippet = %q, want highlighted", snippet)
	}
}

func TestParseSearchHitReturnsIDOnlyWhenNoFormatted(t *testing.T) {
	hit := map[string]json.RawMessage{
		"id": json.RawMessage(`"t2"`),
	}
	id, snippet := parseSearchHit(hit, "title")
	if id != "t2" || snippet != "" {
		t.Fatalf("got id=%q snippet=%q; want id=t2 snippet=empty", id, snippet)
	}
}

func TestParseSearchHitExcludesSnippetWithoutMarkTag(t *testing.T) {
	hit := map[string]json.RawMessage{
		"id":         json.RawMessage(`"t3"`),
		"_formatted": json.RawMessage(`{"title": "plain text without marks"}`),
	}
	id, snippet := parseSearchHit(hit, "title")
	if id != "t3" {
		t.Fatalf("id = %q, want t3", id)
	}
	if snippet != "" {
		t.Fatalf("expected snippet empty when no mark tag, got %q", snippet)
	}
}

func TestParseSearchHitSkipsHitsWithMissingOrBadID(t *testing.T) {
	cases := []map[string]json.RawMessage{
		{"title": json.RawMessage(`"noid"`)}, // no id field at all
		{"id": json.RawMessage(`""`)},        // empty id
		{"id": json.RawMessage(`{invalid}`)}, // malformed json for id
	}
	for i, hit := range cases {
		id, snippet := parseSearchHit(hit, "title")
		if id != "" || snippet != "" {
			t.Fatalf("case %d: expected empty id and snippet; got id=%q snippet=%q", i, id, snippet)
		}
	}
}

func TestParseSearchHitAttributeIsolation(t *testing.T) {
	// The highlight attr is "title"; snippet should be extracted from title, not artistName.
	hit := map[string]json.RawMessage{
		"id":         json.RawMessage(`"t4"`),
		"_formatted": json.RawMessage(`{"title": "<mark>A</mark>", "artistName": "<mark>B</mark>"}`),
	}
	_, snippet := parseSearchHit(hit, "title")
	if snippet != "<mark>A</mark>" {
		t.Fatalf("expected title snippet, got %q", snippet)
	}
}
