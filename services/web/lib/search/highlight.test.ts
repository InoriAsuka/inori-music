import { describe, expect, it } from "vitest";
import { parseHighlight } from "./highlight";

describe("parseHighlight", () => {
  it("returns null for null/undefined input", () => {
    expect(parseHighlight(null)).toBeNull();
    expect(parseHighlight(undefined)).toBeNull();
  });

  it("returns null for empty string", () => {
    expect(parseHighlight("")).toBeNull();
  });

  it("returns null when there is no <mark> tag", () => {
    expect(parseHighlight("plain text")).toBeNull();
  });

  it("parses a single marked segment", () => {
    expect(parseHighlight("hello <mark>w</mark>orld")).toEqual([
      { text: "hello ", marked: false },
      { text: "w", marked: true },
      { text: "orld", marked: false },
    ]);
  });

  it("parses multiple marked segments", () => {
    expect(parseHighlight("<mark>foo</mark> bar <mark>baz</mark>")).toEqual([
      { text: "foo", marked: true },
      { text: " bar ", marked: false },
      { text: "baz", marked: true },
    ]);
  });

  it("handles a mark at the very start with no leading text", () => {
    expect(parseHighlight("<mark>lead</mark>ing")).toEqual([
      { text: "lead", marked: true },
      { text: "ing", marked: false },
    ]);
  });

  it("handles a mark at the very end with no trailing text", () => {
    expect(parseHighlight("trail<mark>ing</mark>")).toEqual([
      { text: "trail", marked: false },
      { text: "ing", marked: true },
    ]);
  });

  it("drops empty marked segments", () => {
    expect(parseHighlight("a<mark></mark>b")).toEqual([
      { text: "a", marked: false },
      { text: "b", marked: false },
    ]);
  });

  it("emits the remainder as plain text on a mismatched opening tag", () => {
    expect(parseHighlight("before <mark>unterminated")).toEqual([
      { text: "before ", marked: false },
      { text: "<mark>unterminated", marked: false },
    ]);
  });
});
