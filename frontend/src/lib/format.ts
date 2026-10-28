import { currentLang, locale, t } from "./i18n";

export function num(v: number | null | undefined, digits = 0) {
  if (v === null || v === undefined || !Number.isFinite(v)) return "—";
  return v.toLocaleString(locale(), { minimumFractionDigits: digits, maximumFractionDigits: digits });
}

export function share(v: number | null | undefined, digits = 0) {
  return `${num(v, digits)}${currentLang() === "es" ? " %" : "%"}`;
}

export function pct(v: number | null | undefined, digits = 0) {
  if (v === null || v === undefined || !Number.isFinite(v)) return "—";
  return `${v > 0 ? "+" : v < 0 ? "−" : ""}${share(Math.abs(v), digits)}`;
}

export function shortDay(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString(locale(), { day: "numeric", month: "short" }).replace(".", "");
}

export function when(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString(locale(), { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" });
}

type Metric = { label: string; unit: string; better: 1 | -1; digits: number };

export function metricMeta(key: string): Metric | undefined {
  switch (key) {
    case "fps":
      return { label: t("FPS medio"), unit: "FPS", better: 1, digits: 0 };
    case "low1":
      return { label: t("1 % low"), unit: "FPS", better: 1, digits: 0 };
    case "low01":
      return { label: t("0,1 % low"), unit: "FPS", better: 1, digits: 0 };
    case "stutterMin":
      return { label: t("Tirones"), unit: t("/min"), better: -1, digits: 0 };
    case "latMs":
      return { label: t("Latencia de render"), unit: "ms", better: -1, digits: 1 };
    case "inputLatency":
      return { label: t("Latencia de entrada"), unit: "ms", better: -1, digits: 1 };
    case "ftStdev":
      return { label: t("Variación del frametime"), unit: "ms", better: -1, digits: 2 };
    case "rtt":
      return { label: t("Ping real"), unit: "ms", better: -1, digits: 1 };
    case "jitter":
      return { label: "Jitter", unit: "ms", better: -1, digits: 1 };
    case "lossDown":
      return { label: t("Pérdida"), unit: "%", better: -1, digits: 2 };
    case "maxGap":
      return { label: t("Mayor corte"), unit: "ms", better: -1, digits: 0 };
    case "hardFaults":
      return { label: t("Lecturas de paginación"), unit: "/s", better: -1, digits: 0 };
    case "ramRoblox":
      return { label: t("RAM de Roblox"), unit: "MB", better: -1, digits: 0 };
    case "cpuBusy":
      return { label: t("CPU por frame"), unit: "ms", better: -1, digits: 2 };
    case "gpuBusy":
      return { label: t("GPU por frame"), unit: "ms", better: -1, digits: 2 };
    case "cpuRoblox":
      return { label: t("CPU de Roblox"), unit: "%", better: -1, digits: 0 };
    default:
      return undefined;
  }
}

export function improvement(key: string, deltaPct: number) {
  const m = metricMeta(key);
  if (!m) return 0;
  return deltaPct * m.better;
}
