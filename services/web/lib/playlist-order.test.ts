import { describe, expect, it } from "vitest";
import { firstIndexOf, moveAt, moveId, occurrenceAt, removeAt } from "./playlist-order";

describe("playlist occurrence ordering", () => {
  it("moves one duplicate occurrence without collapsing duplicates", () => {
    expect(moveId(["a", "b", "a", "c"], 2, 1)).toEqual(["a", "a", "b", "c"]);
  });

  it("moves occurrence rows while preserving their stable identities", () => {
    const rows = [
      { uid: "a#0", id: "a" },
      { uid: "b#0", id: "b" },
      { uid: "a#1", id: "a" },
    ];
    expect(moveAt(rows, 2, 0).map((row) => row.uid)).toEqual(["a#1", "a#0", "b#0"]);
  });

  it("removes only the requested occurrence", () => {
    expect(removeAt(["a", "b", "a", "c"], 2)).toEqual(["a", "b", "c"]);
    expect(removeAt(["a", "b", "a", "c"], 0)).toEqual(["b", "a", "c"]);
  });

  it("identifies whether DELETE can remove the selected occurrence", () => {
    const ids = ["a", "b", "a"];
    expect(firstIndexOf(ids, "a")).toBe(0);
    expect(firstIndexOf(ids, "missing")).toBe(-1);
  });

  it("describes the selected duplicate occurrence for confirmation", () => {
    const ids = ["a", "b", "a", "a"];
    expect(occurrenceAt(ids, 0)).toEqual({ occurrence: 1, total: 3 });
    expect(occurrenceAt(ids, 2)).toEqual({ occurrence: 2, total: 3 });
    expect(occurrenceAt(ids, 3)).toEqual({ occurrence: 3, total: 3 });
    expect(occurrenceAt(ids, 1)).toEqual({ occurrence: 1, total: 1 });
    expect(occurrenceAt(ids, 4)).toBeNull();
  });

  it("returns copies and leaves input unchanged for invalid positions", () => {
    const ids = ["a", "b"];
    expect(moveId(ids, -1, 1)).toEqual(ids);
    expect(removeAt(ids, 4)).toEqual(ids);
    expect(moveId(ids, 0, 0)).not.toBe(ids);
    expect(ids).toEqual(["a", "b"]);
  });
});
