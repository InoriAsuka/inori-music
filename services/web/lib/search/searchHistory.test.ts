import { describe, expect, it } from "vitest";
import { addEntry, removeEntry } from "./searchHistory";

describe("addEntry", () => {
  it("adds a query to the front of an empty list", () => {
    expect(addEntry([], "hello")).toEqual(["hello"]);
  });

  it("trims whitespace before adding", () => {
    expect(addEntry([], "  spaced  ")).toEqual(["spaced"]);
  });

  it("ignores a blank query", () => {
    expect(addEntry(["a"], "   ")).toEqual(["a"]);
  });

  it("dedups and moves an existing query to the front", () => {
    expect(addEntry(["a", "b"], "a")).toEqual(["a", "b"]);
    expect(addEntry(["b", "a"], "a")).toEqual(["a", "b"]);
  });

  it("caps the result at 20 entries, dropping the oldest", () => {
    const entries = Array.from({ length: 20 }, (_, i) => `q${i}`);
    const next = addEntry(entries, "new");
    expect(next).toHaveLength(20);
    expect(next[0]).toBe("new");
    expect(next).not.toContain("q19");
  });
});

describe("removeEntry", () => {
  it("removes a matching entry", () => {
    expect(removeEntry(["a", "b", "c"], "b")).toEqual(["a", "c"]);
  });

  it("is a no-op when the entry isn't present", () => {
    expect(removeEntry(["a", "b"], "z")).toEqual(["a", "b"]);
  });
});
