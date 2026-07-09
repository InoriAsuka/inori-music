import { describe, expect, it } from "vitest";
import { cn, formatDuration, truncate } from "./utils";

describe("cn", () => {
  it("merges plain class strings", () => {
    expect(cn("a", "b")).toBe("a b");
  });

  it("drops falsy values", () => {
    expect(cn("a", false && "b", undefined, null, "c")).toBe("a c");
  });

  it("resolves conflicting Tailwind classes to the last one (twMerge)", () => {
    expect(cn("px-2", "px-4")).toBe("px-4");
  });
});

describe("formatDuration", () => {
  it("formats whole minutes and seconds as mm:ss", () => {
    expect(formatDuration(65)).toBe("1:05");
  });

  it("pads single-digit seconds with a leading zero", () => {
    expect(formatDuration(9)).toBe("0:09");
  });

  it("formats zero as 0:00", () => {
    expect(formatDuration(0)).toBe("0:00");
  });

  it("does not pad the minutes component", () => {
    expect(formatDuration(600)).toBe("10:00");
  });

  it("truncates fractional seconds", () => {
    expect(formatDuration(90.9)).toBe("1:30");
  });
});

describe("truncate", () => {
  it("returns the string unchanged when within maxLen", () => {
    expect(truncate("hello", 10)).toBe("hello");
  });

  it("truncates and appends an ellipsis when over maxLen", () => {
    expect(truncate("hello world", 8)).toBe("hello w…");
  });

  it("returns the string unchanged at exactly maxLen", () => {
    expect(truncate("hello", 5)).toBe("hello");
  });
});
