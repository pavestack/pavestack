/** Deterministic pseudo-random generator so sample metrics are stable across renders. */
function mulberry32(seed: number) {
  return function () {
    seed |= 0;
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function seedFromString(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = (Math.imul(h, 31) + s.charCodeAt(i)) | 0;
  return h;
}

export type GoldenSignals = {
  requestRate: number[];
  errorRate: number[];
  latencyP99: number[];
  deployMarkers: number[];
  errorSpikeMarkers: number[];
  currentRequestRate: number;
  currentErrorRate: number;
  currentLatencyP99: number;
  sloBurnRate: number;
};

const POINTS = 24;

/** Illustrative golden-signal sample series for a service — no live Prometheus source exists yet. */
export function generateSampleMetrics(serviceName: string): GoldenSignals {
  const rand = mulberry32(seedFromString(serviceName));

  const requestRate: number[] = [];
  const errorRate: number[] = [];
  const latencyP99: number[] = [];
  const deployMarkers: number[] = [];
  const errorSpikeMarkers: number[] = [];

  let rps = 40 + rand() * 200;
  let err = rand() * 1.2;
  let latency = 80 + rand() * 120;

  for (let i = 0; i < POINTS; i++) {
    rps += (rand() - 0.5) * 15;
    err += (rand() - 0.5) * 0.3;
    latency += (rand() - 0.5) * 20;

    if (rand() > 0.88) {
      deployMarkers.push(i);
      // Deploys occasionally correlate with a brief error/latency bump — sample data only.
      if (rand() > 0.5) {
        err += 1.5 + rand();
        latency += 40;
        errorSpikeMarkers.push(i);
      }
    }

    requestRate.push(Math.max(0, rps));
    errorRate.push(Math.max(0, err));
    latencyP99.push(Math.max(10, latency));
  }

  return {
    requestRate,
    errorRate,
    latencyP99,
    deployMarkers,
    errorSpikeMarkers,
    currentRequestRate: Math.round(requestRate[requestRate.length - 1]),
    currentErrorRate: Number(errorRate[errorRate.length - 1].toFixed(2)),
    currentLatencyP99: Math.round(latencyP99[latencyP99.length - 1]),
    sloBurnRate: Number((0.2 + rand() * 1.6).toFixed(2)),
  };
}
