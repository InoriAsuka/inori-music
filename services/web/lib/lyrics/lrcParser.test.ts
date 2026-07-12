import { describe, expect, it } from "vitest";
import { parseLrc } from "./lrcParser";

describe("parseLrc", () => {
  it("parses standard line-level [mm:ss.xx] timestamps", () => {
    const content = "[00:12.50]First line\n[00:15.00]Second line";
    const lines = parseLrc(content);
    expect(lines).toEqual([
      { timestampMs: 12_500, text: "First line", words: undefined },
      { timestampMs: 15_000, text: "Second line", words: undefined },
    ]);
  });

  it("parses [mm:ss.xxx] millisecond-precision timestamps", () => {
    const lines = parseLrc("[00:01.234]Hi");
    expect(lines[0].timestampMs).toBe(1234);
  });

  it("supports colon as the decimal separator", () => {
    const lines = parseLrc("[00:01:50]Hi");
    expect(lines[0].timestampMs).toBe(1500);
  });

  it("sorts lines by timestamp regardless of source order", () => {
    const content = "[00:20.00]Second\n[00:10.00]First";
    const lines = parseLrc(content);
    expect(lines.map((l) => l.text)).toEqual(["First", "Second"]);
  });

  it("skips metadata-only tags like [ti:], [ar:], [al:]", () => {
    const content = "[ti:Song Title]\n[ar:Artist]\n[00:05.00]Actual lyric";
    const lines = parseLrc(content);
    expect(lines).toHaveLength(1);
    expect(lines[0].text).toBe("Actual lyric");
  });

  it("skips lines with an empty timestamp payload", () => {
    const content = "[00:05.00]\n[00:06.00]Real line";
    const lines = parseLrc(content);
    expect(lines).toHaveLength(1);
    expect(lines[0].text).toBe("Real line");
  });

  it("ignores lines with no timestamp tag at all", () => {
    const content = "not a lyric line\n[00:05.00]Real line";
    const lines = parseLrc(content);
    expect(lines).toHaveLength(1);
  });

  it("returns an empty array for empty content", () => {
    expect(parseLrc("")).toEqual([]);
  });

  it("parses word-level karaoke tags into words[]", () => {
    const content = "[00:10.00]<00:10.00>Hel<00:10.30>lo <00:10.60>world";
    const lines = parseLrc(content);
    expect(lines).toHaveLength(1);
    expect(lines[0].text).toBe("Hello world");
    expect(lines[0].words).toEqual([
      { offsetMs: 10_000, text: "Hel" },
      { offsetMs: 10_300, text: "lo " },
      { offsetMs: 10_600, text: "world" },
    ]);
  });

  it("keeps text preceding the first inline tag as a fragment timed at the line timestamp", () => {
    const content = "[00:10.00]pre <00:10.50>tagged";
    const lines = parseLrc(content);
    expect(lines[0].words).toEqual([
      { offsetMs: 10_000, text: "pre " },
      { offsetMs: 10_500, text: "tagged" },
    ]);
  });

  it("leaves words undefined for lines without inline tags", () => {
    const lines = parseLrc("[00:10.00]plain line");
    expect(lines[0].words).toBeUndefined();
  });

  it("drops adjacent inline tags that produce an empty fragment", () => {
    const content = "[00:10.00]<00:10.00><00:10.10>text";
    const lines = parseLrc(content);
    expect(lines[0].words).toEqual([{ offsetMs: 10_100, text: "text" }]);
  });

  it("trims trailing/leading whitespace from raw lines before matching", () => {
    const lines = parseLrc("   [00:05.00]  spaced  \n");
    expect(lines[0].text).toBe("spaced");
  });
});
