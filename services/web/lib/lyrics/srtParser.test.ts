import { describe, expect, it } from "vitest";
import { parseSrt } from "./srtParser";

describe("parseSrt", () => {
  it("parses a standard SRT block", () => {
    const content = "1\n00:00:12,500 --> 00:00:15,000\nFirst line";
    const lines = parseSrt(content);
    expect(lines).toEqual([{ timestampMs: 12_500, text: "First line" }]);
  });

  it("supports a dot as the millisecond separator", () => {
    const lines = parseSrt("1\n00:00:12.500 --> 00:00:15.000\nHi");
    expect(lines[0].timestampMs).toBe(12_500);
  });

  it("joins multi-line subtitle text with a space", () => {
    const content = "1\n00:00:01,000 --> 00:00:02,000\nLine one\nLine two";
    const lines = parseSrt(content);
    expect(lines[0].text).toBe("Line one Line two");
  });

  it("parses multiple blocks separated by blank lines and sorts by timestamp", () => {
    const content = ["2", "00:00:20,000 --> 00:00:22,000", "Second", "", "1", "00:00:10,000 --> 00:00:12,000", "First"].join(
      "\n"
    );
    const lines = parseSrt(content);
    expect(lines.map((l) => l.text)).toEqual(["First", "Second"]);
  });

  it("includes the hours component in the timestamp", () => {
    const lines = parseSrt("1\n01:00:00,000 --> 01:00:01,000\nHour mark");
    expect(lines[0].timestampMs).toBe(3_600_000);
  });

  it("skips malformed blocks with no time range", () => {
    const content = "1\nnot a time range\ntext";
    expect(parseSrt(content)).toEqual([]);
  });

  it("skips blocks with empty text", () => {
    const content = "1\n00:00:01,000 --> 00:00:02,000\n";
    expect(parseSrt(content)).toEqual([]);
  });

  it("returns an empty array for empty content", () => {
    expect(parseSrt("")).toEqual([]);
  });
});
