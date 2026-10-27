const es = "es-CO";

export function num(v: number | null | undefined, digits = 0) {
  if (v === null || v === undefined || !Number.isFinite(v)) return "—";
  return v.toLocaleString(es, { minimumFractionDigits: digits, maximumFractionDigits: digits });
}

export function pct(v: number | null | undefined, digits = 0) {
  if (v === null || v === undefined || !Number.isFinite(v)) return "—";
  const s = Math.abs(v).toLocaleString(es, { minimumFractionDigits: digits, maximumFractionDigits: digits });
  return `${v > 0 ? "+" : v < 0 ? "−" : ""}${s} %`;
}

export function when(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString(es, { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" });
}

export const metricMeta: Record<string, { label: string; unit: string; better: 1 | -1; digits: number }> = {
  fps: { label: "FPS medio", unit: "FPS", better: 1, digits: 0 },
  low1: { label: "1 % low", unit: "FPS", better: 1, digits: 0 },
  low01: { label: "0,1 % low", unit: "FPS", better: 1, digits: 0 },
  stutterMin: { label: "Tirones", unit: "/min", better: -1, digits: 0 },
  latMs: { label: "Latencia de render", unit: "ms", better: -1, digits: 1 },
  inputLatency: { label: "Latencia de entrada", unit: "ms", better: -1, digits: 1 },
  ftStdev: { label: "Variación del frametime", unit: "ms", better: -1, digits: 2 },
  rtt: { label: "Ping real", unit: "ms", better: -1, digits: 1 },
  jitter: { label: "Jitter", unit: "ms", better: -1, digits: 1 },
  lossDown: { label: "Pérdida", unit: "%", better: -1, digits: 2 },
  maxGap: { label: "Mayor corte", unit: "ms", better: -1, digits: 0 },
  hardFaults: { label: "Lecturas de paginación", unit: "/s", better: -1, digits: 0 },
  ramRoblox: { label: "RAM de Roblox", unit: "MB", better: -1, digits: 0 },
  cpuBusy: { label: "CPU por frame", unit: "ms", better: -1, digits: 2 },
  gpuBusy: { label: "GPU por frame", unit: "ms", better: -1, digits: 2 },
  cpuRoblox: { label: "CPU de Roblox", unit: "%", better: -1, digits: 0 },
};

export function improvement(key: string, deltaPct: number) {
  const m = metricMeta[key];
  if (!m) return 0;
  return deltaPct * m.better;
}
