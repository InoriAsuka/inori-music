import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

/**
 * Unit tests for audioGraph.ts — the WebAudio gain/EQ pipeline wrapping an
 * HTMLAudioElement. Vitest runs in a `node` environment with no real
 * AudioContext, so we install a minimal fake `window.AudioContext` whose
 * nodes are plain objects with `vi.fn()` methods. That's enough to assert
 * the graph's wiring order, the gain/EQ math it schedules, and the
 * CORS/unsupported-browser fallback without needing a browser.
 *
 * Module state (`sharedContext`, `resumeListenersInstalled`) is file-level
 * in audioGraph.ts, so each test that cares about a fresh context calls
 * `vi.resetModules()` and re-imports — the same isolation approach as
 * lib/api/token.test.ts.
 */

type FakeParam = {
  value: number;
  cancelScheduledValues: ReturnType<typeof vi.fn>;
  setValueAtTime: ReturnType<typeof vi.fn>;
  linearRampToValueAtTime: ReturnType<typeof vi.fn>;
};

function makeParam(initial: number): FakeParam {
  return {
    value: initial,
    cancelScheduledValues: vi.fn(),
    setValueAtTime: vi.fn(),
    linearRampToValueAtTime: vi.fn(),
  };
}

function makeFilterNode() {
  return {
    type: "",
    frequency: { value: 0 },
    Q: { value: 0 },
    gain: makeParam(0),
    connect: vi.fn(),
    disconnect: vi.fn(),
  };
}

function makeGainNode() {
  return {
    gain: makeParam(0),
    connect: vi.fn(),
    disconnect: vi.fn(),
  };
}

function makeSourceNode() {
  return { connect: vi.fn(), disconnect: vi.fn() };
}

interface FakeAudioContextOptions {
  /** Install the constructor under `webkitAudioContext` instead of `AudioContext`. */
  webkitOnly?: boolean;
  /** Make `createMediaElementSource` throw, simulating a CORS-tainted/already-connected element. */
  throwOnCreateSource?: boolean;
}

function installFakeAudioContext(options: FakeAudioContextOptions = {}) {
  const filters: ReturnType<typeof makeFilterNode>[] = [];
  const gainNodes: ReturnType<typeof makeGainNode>[] = [];
  const sources: ReturnType<typeof makeSourceNode>[] = [];
  const contexts: Array<{
    currentTime: number;
    state: string;
    destination: { label: string };
    resume: ReturnType<typeof vi.fn>;
  }> = [];

  const AudioContextCtor = vi.fn(function (this: Record<string, unknown>) {
    this.currentTime = 0;
    this.state = "running";
    this.destination = { label: "destination" };
    this.resume = vi.fn(() => Promise.resolve());
    this.createMediaElementSource = vi.fn(() => {
      if (options.throwOnCreateSource) throw new Error("CORS tainted");
      const s = makeSourceNode();
      sources.push(s);
      return s;
    });
    this.createBiquadFilter = vi.fn(() => {
      const f = makeFilterNode();
      filters.push(f);
      return f;
    });
    this.createGain = vi.fn(() => {
      const g = makeGainNode();
      gainNodes.push(g);
      return g;
    });
    contexts.push(this as unknown as (typeof contexts)[number]);
  });

  const addEventListener = vi.fn();
  const fakeWindow: Record<string, unknown> = { addEventListener };
  fakeWindow[options.webkitOnly ? "webkitAudioContext" : "AudioContext"] = AudioContextCtor;

  vi.stubGlobal("window", fakeWindow);

  return { AudioContextCtor, fakeWindow, addEventListener, filters, gainNodes, sources, contexts };
}

function makeAudioElement(): HTMLAudioElement {
  return { crossOrigin: "" } as unknown as HTMLAudioElement;
}

beforeEach(() => {
  vi.resetModules();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("createAudioGraph — no AudioContext available (SSR / unsupported browser)", () => {
  it("returns an inert passthrough node without throwing, still tagging the element for CORS", async () => {
    vi.stubGlobal("window", undefined);
    const { createAudioGraph } = await import("./audioGraph");
    const audio = makeAudioElement();

    const node = createAudioGraph(audio);

    expect(node.active).toBe(false);
    expect(node.gainNode).toBeNull();
    expect(node.eqFilters).toEqual([]);
    expect(audio.crossOrigin).toBe("anonymous");
    expect(() => node.setGain(1)).not.toThrow();
    expect(() => node.rampGain(1, 1)).not.toThrow();
    expect(() => node.setEqBand(0, 1)).not.toThrow();
    expect(() => node.disconnect()).not.toThrow();
  });

  it("falls back when window exists but neither AudioContext nor webkitAudioContext is defined", async () => {
    vi.stubGlobal("window", { addEventListener: vi.fn() });
    const { createAudioGraph } = await import("./audioGraph");

    const node = createAudioGraph(makeAudioElement());

    expect(node.active).toBe(false);
  });
});

describe("createAudioGraph — successful graph creation", () => {
  it("sets crossOrigin=anonymous on the element", async () => {
    installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const audio = makeAudioElement();

    createAudioGraph(audio);

    expect(audio.crossOrigin).toBe("anonymous");
  });

  it("reports active with a real gain node and 10 EQ filters", async () => {
    installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");

    const node = createAudioGraph(makeAudioElement());

    expect(node.active).toBe(true);
    expect(node.gainNode).not.toBeNull();
    expect(node.eqFilters).toHaveLength(10);
  });

  it("creates the 10 bands at the documented frequencies with peaking type, Q=1.4, 0dB initial gain", async () => {
    const { filters } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");

    createAudioGraph(makeAudioElement());

    expect(filters.map((f) => f.frequency.value)).toEqual([31, 62, 125, 250, 500, 1000, 2000, 4000, 8000, 16000]);
    for (const f of filters) {
      expect(f.type).toBe("peaking");
      expect(f.Q.value).toBe(1.4);
      expect(f.gain.value).toBe(0);
    }
  });

  it("wires source -> filter[0] -> ... -> filter[9] -> gain -> destination", async () => {
    const { filters, gainNodes, sources, contexts } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");

    createAudioGraph(makeAudioElement());

    expect(sources[0].connect).toHaveBeenCalledWith(filters[0]);
    for (let i = 0; i < filters.length - 1; i++) {
      expect(filters[i].connect).toHaveBeenCalledWith(filters[i + 1]);
    }
    expect(filters[9].connect).toHaveBeenCalledWith(gainNodes[0]);
    expect(gainNodes[0].connect).toHaveBeenCalledWith(contexts[0].destination);
  });

  it("initializes gain to unity (1.0)", async () => {
    const { gainNodes } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");

    createAudioGraph(makeAudioElement());

    expect(gainNodes[0].gain.value).toBe(1.0);
  });

  it("falls back to webkitAudioContext when AudioContext is unavailable", async () => {
    const { filters } = installFakeAudioContext({ webkitOnly: true });
    const { createAudioGraph } = await import("./audioGraph");

    const node = createAudioGraph(makeAudioElement());

    expect(node.active).toBe(true);
    expect(filters).toHaveLength(10);
  });

  it("falls back to a passthrough node when createMediaElementSource throws (CORS-tainted element)", async () => {
    installFakeAudioContext({ throwOnCreateSource: true });
    const { createAudioGraph } = await import("./audioGraph");

    const node = createAudioGraph(makeAudioElement());

    expect(node.active).toBe(false);
    expect(node.gainNode).toBeNull();
    expect(node.eqFilters).toEqual([]);
    expect(() => node.setGain(1)).not.toThrow();
    expect(() => node.disconnect()).not.toThrow();
  });
});

describe("AudioGraphNode.setGain", () => {
  it("cancels scheduled changes and sets the value immediately at the current time", async () => {
    const { gainNodes, contexts } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());
    contexts[0].currentTime = 5;

    node.setGain(0.5);

    const gain = gainNodes[0].gain;
    expect(gain.cancelScheduledValues).toHaveBeenCalledWith(5);
    expect(gain.setValueAtTime).toHaveBeenCalledWith(0.5, 5);
  });

  it("is a no-op on the inert fallback node", async () => {
    vi.stubGlobal("window", undefined);
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());

    expect(() => node.setGain(0.5)).not.toThrow();
  });
});

describe("AudioGraphNode.rampGain", () => {
  it("ramps linearly from the current value to the target over the given duration", async () => {
    const { gainNodes, contexts } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());
    const gain = gainNodes[0].gain;
    contexts[0].currentTime = 10;
    gain.value = 0.2;

    node.rampGain(1, 2);

    expect(gain.cancelScheduledValues).toHaveBeenCalledWith(10);
    expect(gain.setValueAtTime).toHaveBeenCalledWith(0.2, 10);
    expect(gain.linearRampToValueAtTime).toHaveBeenCalledWith(1, 12);
  });
});

describe("AudioGraphNode.setEqBand", () => {
  it("updates only the targeted band at the current time", async () => {
    const { filters, contexts } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());
    contexts[0].currentTime = 1;

    node.setEqBand(3, -4);

    expect(filters[3].gain.cancelScheduledValues).toHaveBeenCalledWith(1);
    expect(filters[3].gain.setValueAtTime).toHaveBeenCalledWith(-4, 1);
    for (const [i, f] of filters.entries()) {
      if (i !== 3) expect(f.gain.setValueAtTime).not.toHaveBeenCalled();
    }
  });

  it("ignores out-of-range indices instead of throwing", async () => {
    const { filters } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());

    expect(() => node.setEqBand(-1, 3)).not.toThrow();
    expect(() => node.setEqBand(10, 3)).not.toThrow();
    for (const f of filters) {
      expect(f.gain.setValueAtTime).not.toHaveBeenCalled();
    }
  });
});

describe("AudioGraphNode.disconnect", () => {
  it("tears down the source, every EQ filter, and the gain node", async () => {
    const { filters, gainNodes, sources } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());

    node.disconnect();

    expect(sources[0].disconnect).toHaveBeenCalledTimes(1);
    for (const f of filters) expect(f.disconnect).toHaveBeenCalledTimes(1);
    expect(gainNodes[0].disconnect).toHaveBeenCalledTimes(1);
  });

  it("does not throw if a node was already disconnected", async () => {
    const { sources } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    const node = createAudioGraph(makeAudioElement());
    sources[0].disconnect = vi.fn(() => {
      throw new Error("already disconnected");
    });

    expect(() => node.disconnect()).not.toThrow();
  });
});

describe("shared AudioContext reuse across slots", () => {
  it("creates the underlying AudioContext only once for multiple graph nodes", async () => {
    const { AudioContextCtor } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");

    createAudioGraph(makeAudioElement());
    createAudioGraph(makeAudioElement());

    expect(AudioContextCtor).toHaveBeenCalledTimes(1);
  });
});

describe("resume-on-gesture installation", () => {
  it("registers pointerdown, keydown, and touchstart listeners exactly once, even across multiple graphs", async () => {
    const { addEventListener } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");

    createAudioGraph(makeAudioElement());
    createAudioGraph(makeAudioElement());

    expect(addEventListener).toHaveBeenCalledTimes(3);
    const eventNames = addEventListener.mock.calls.map((call) => call[0]).sort();
    expect(eventNames).toEqual(["keydown", "pointerdown", "touchstart"]);
  });

  it("resumes a suspended context on the first gesture", async () => {
    const { addEventListener, contexts } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    createAudioGraph(makeAudioElement());
    contexts[0].state = "suspended";

    const resumeHandler = addEventListener.mock.calls[0][1] as () => void;
    resumeHandler();

    expect(contexts[0].resume).toHaveBeenCalledTimes(1);
  });

  it("does not call resume when the context is already running", async () => {
    const { addEventListener, contexts } = installFakeAudioContext();
    const { createAudioGraph } = await import("./audioGraph");
    createAudioGraph(makeAudioElement());
    contexts[0].state = "running";

    const resumeHandler = addEventListener.mock.calls[0][1] as () => void;
    resumeHandler();

    expect(contexts[0].resume).not.toHaveBeenCalled();
  });
});
