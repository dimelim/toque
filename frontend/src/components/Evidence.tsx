import { useLayoutEffect, useRef, useState } from "react";
import { metricMeta, num, pct, when } from "@/lib/format";

export interface StripPoint {
  arm: "A" | "B";
  value: number;
  at: string;
  index: number;
}

function useWidth<T extends HTMLElement>() {
  const ref = useRef<T>(null);
  const [width, setWidth] = useState(0);
  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    const ro = new ResizeObserver(([entry]) => setWidth(entry.contentRect.width));
    ro.observe(el);
    return () => ro.disconnect();
  }, []);
  return [ref, width] as const;
}

export function median(values: number[]) {
  if (values.length === 0) return NaN;
  const s = [...values].sort((a, b) => a - b);
  const m = Math.floor(s.length / 2);
  return s.length % 2 ? s[m] : (s[m - 1] + s[m]) / 2;
}

function lanes(xs: number[], gap: number) {
  const order = xs.map((x, i) => [x, i] as const).sort((a, b) => a[0] - b[0]);
  const offsets = [0, -7, 7, -14, 14];
  const last = offsets.map(() => -Infinity);
  const out: number[] = new Array(xs.length).fill(0);
  for (const [x, i] of order) {
    let lane = offsets.findIndex((_, k) => x - last[k] >= gap);
    if (lane < 0) lane = 0;
    last[lane] = x;
    out[i] = offsets[lane];
  }
  return out;
}

const ROW_A = 28;
const ROW_B = 76;
const HEIGHT = 112;
const PAD = 14;

export function EvidenceStrip({ metric, points }: { metric: string; points: StripPoint[] }) {
  const [ref, width] = useWidth<HTMLDivElement>();
  const [hover, setHover] = useState<number | null>(null);
  const meta = metricMeta[metric];
  const values = points.map((p) => p.value);
  const lo = Math.min(...values);
  const hi = Math.max(...values);
  const span = hi - lo || Math.abs(hi) * 0.05 || 1;
  const x0 = lo - span * 0.08;
  const x1 = hi + span * 0.08;
  const sx = (v: number) => PAD + ((v - x0) / (x1 - x0)) * Math.max(1, width - 2 * PAD);
  const xs = points.map((p) => sx(p.value));
  const offA = lanes(xs.filter((_, i) => points[i].arm === "A"), 9);
  const offB = lanes(xs.filter((_, i) => points[i].arm === "B"), 9);
  let ia = 0;
  let ib = 0;
  const ys = points.map((p) => (p.arm === "A" ? ROW_A + offA[ia++] : ROW_B + offB[ib++]));
  const medA = median(points.filter((p) => p.arm === "A").map((p) => p.value));
  const medB = median(points.filter((p) => p.arm === "B").map((p) => p.value));
  const h = hover !== null ? points[hover] : null;
  const digits = (meta?.digits ?? 1) + 1;
  return (
    <div className="flex w-full gap-3">
      <div className="relative w-[104px] shrink-0" style={{ height: HEIGHT }}>
        {[
          { y: ROW_A, label: "Sin el cambio", value: medA, dot: "bg-before", ink: "text-ink-2" },
          { y: ROW_B, label: "Con el cambio", value: medB, dot: "bg-volt", ink: "text-ink" },
        ].map((r) => (
          <div key={r.label} className="absolute left-0 -translate-y-1/2" style={{ top: r.y }}>
            <p className="flex items-center gap-1.5 text-caption-1-medium text-ink-3">
              <span className={`size-2 rounded-full ${r.dot}`} />
              {r.label}
            </p>
            <p className={`num mt-0.5 text-[16px] leading-none font-semibold ${r.ink}`}>
              {num(r.value, digits)} <span className="font-sans text-[11px] font-medium text-ink-4">{meta?.unit}</span>
            </p>
          </div>
        ))}
      </div>
    <div ref={ref} className="relative min-w-0 flex-1" style={{ height: HEIGHT }}>
      {width > 0 ? (
        <svg width={width} height={HEIGHT} role="img" aria-label={`${meta?.label ?? metric}, mediana sin el cambio ${num(medA, meta?.digits ?? 1)} y con el cambio ${num(medB, meta?.digits ?? 1)} ${meta?.unit ?? ""}`}>
          <line x1={PAD} x2={width - PAD} y1={HEIGHT - 18} y2={HEIGHT - 18} stroke="var(--color-line-strong)" strokeWidth={1} />
          {[medA, medB].map((m, k) =>
            Number.isFinite(m) ? (
              <line key={k} x1={sx(m)} x2={sx(m)} y1={(k ? ROW_B : ROW_A) - 20} y2={(k ? ROW_B : ROW_A) + 20} stroke="var(--color-ink-3)" strokeWidth={2} strokeLinecap="round" />
            ) : null,
          )}
          {points.map((p, i) => (
            <circle
              key={i}
              cx={xs[i]}
              cy={ys[i]}
              r={hover === i ? 5.5 : 4}
              fill={p.arm === "B" ? "var(--color-volt)" : "var(--color-before)"}
              stroke="var(--color-surface-1)"
              strokeWidth={2}
            />
          ))}
          {points.map((_, i) => (
            <circle
              key={`hit-${i}`}
              cx={xs[i]}
              cy={ys[i]}
              r={12}
              fill="transparent"
              onPointerEnter={() => setHover(i)}
              onPointerLeave={() => setHover((cur) => (cur === i ? null : cur))}
            />
          ))}
          <text x={PAD} y={HEIGHT - 2} className="fill-ink-4 text-[11px]" style={{ fontVariantNumeric: "tabular-nums" }}>
            {num(lo, meta?.digits ?? 1)} {meta?.unit}
          </text>
          <text x={width - PAD} y={HEIGHT - 2} textAnchor="end" className="fill-ink-4 text-[11px]" style={{ fontVariantNumeric: "tabular-nums" }}>
            {num(hi, meta?.digits ?? 1)} {meta?.unit}
          </text>
        </svg>
      ) : null}
      {h ? (
        <div
          className="pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-full rounded-lg bg-surface-3 px-2.5 py-1.5 whitespace-nowrap shadow-dropdown"
          style={{ left: xs[hover!], top: ys[hover!] - 10 }}
        >
          <p className="num text-[15px] font-semibold text-ink">
            {num(h.value, (meta?.digits ?? 1) + 1)} <span className="font-sans text-[11px] font-medium text-ink-3">{meta?.unit}</span>
          </p>
          <p className="text-caption-1-medium text-ink-3">
            {h.arm === "B" ? "Con el cambio" : "Sin el cambio"} · ventana {h.index} · {when(h.at)}
          </p>
        </div>
      ) : null}
    </div>
    </div>
  );
}

export function EffectBar({ low, high, value, tone, scale }: { low: number; high: number; value: number; tone: "volt" | "bad" | "neutral"; scale: number }) {
  const [ref, width] = useWidth<HTMLDivElement>();
  const color = tone === "volt" ? "var(--color-volt)" : tone === "bad" ? "var(--color-bad)" : "var(--color-ink-4)";
  const sx = (v: number) => width / 2 + (Math.max(-scale, Math.min(scale, v)) / scale) * (width / 2 - 8);
  return (
    <div ref={ref} className="h-6 w-full">
      {width > 0 ? (
        <svg width={width} height={24} aria-hidden>
          <line x1={width / 2} x2={width / 2} y1={2} y2={22} stroke="var(--color-line-strong)" strokeWidth={1} />
          <line x1={sx(low)} x2={sx(high)} y1={12} y2={12} stroke={color} strokeWidth={2} strokeLinecap="round" />
          <circle cx={sx(value)} cy={12} r={4} fill={color} stroke="var(--color-surface-1)" strokeWidth={2} />
        </svg>
      ) : null}
    </div>
  );
}

export function signedPct(v: number) {
  return pct(v, Math.abs(v) < 10 ? 1 : 0);
}
