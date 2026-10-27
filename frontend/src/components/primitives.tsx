import type { ReactNode } from "react";
import { cx } from "@/utils/cx";
import { improvement, metricMeta, num, pct } from "@/lib/format";

export function Card({ className, children }: { className?: string; children: ReactNode }) {
  return <section className={cx("rounded-2xl bg-background-secondary-default shadow-card", className)}>{children}</section>;
}

export function Eyebrow({ children, className }: { children: ReactNode; className?: string }) {
  return <p className={cx("label-caps text-ink-3", className)}>{children}</p>;
}

export function PageHeader({ eyebrow, title, children }: { eyebrow: string; title: string; children?: ReactNode }) {
  return (
    <header className="flex items-end justify-between gap-6 pb-6">
      <div className="min-w-0">
        <Eyebrow>{eyebrow}</Eyebrow>
        <h1 className="mt-2 text-[28px] leading-[1.1] font-semibold tracking-[-0.02em] text-ink">{title}</h1>
      </div>
      <div className="flex shrink-0 items-center gap-2">{children}</div>
    </header>
  );
}

export function BigNumber({
  value,
  unit,
  label,
  tone = "ink",
  size = "lg",
  digits = 0,
}: {
  value: number | null | undefined;
  unit?: string;
  label: string;
  tone?: "ink" | "volt" | "muted";
  size?: "md" | "lg" | "xl";
  digits?: number;
}) {
  const sizes = { md: "text-[34px]", lg: "text-[46px]", xl: "text-[72px]" };
  const tones = { ink: "text-ink", volt: "text-volt", muted: "text-ink-2" };
  return (
    <div className="min-w-0">
      <p className="text-caption-1-medium text-ink-3">{label}</p>
      <p className={cx("num mt-1 leading-[0.95] font-semibold", sizes[size], tones[tone])}>
        {num(value, digits)}
        {unit && value !== null && value !== undefined ? <span className="ml-1 font-sans text-[13px] font-medium tracking-normal text-ink-3">{unit}</span> : null}
      </p>
    </div>
  );
}

export function Pill({ children, tone = "neutral", className }: { children: ReactNode; tone?: "neutral" | "volt" | "warn" | "bad"; className?: string }) {
  const tones = {
    neutral: "bg-surface-3 text-ink-2",
    volt: "bg-volt-soft text-volt",
    warn: "bg-warn-soft text-warn",
    bad: "bg-bad-soft text-bad",
  };
  return <span className={cx("inline-flex h-6 items-center gap-1 rounded-full px-2.5 text-caption-1-medium whitespace-nowrap", tones[tone], className)}>{children}</span>;
}

export function DeltaChip({ metric, deltaPct, conclusive = true }: { metric: string; deltaPct: number; conclusive?: boolean }) {
  const meta = metricMeta[metric];
  const good = improvement(metric, deltaPct) > 0;
  const tone = !conclusive ? "neutral" : good ? "volt" : "bad";
  return (
    <Pill tone={tone}>
      <span className="text-ink-3">{meta?.label ?? metric}</span>
      <span className="num text-[14px] font-semibold">{pct(deltaPct, Math.abs(deltaPct) < 10 ? 1 : 0)}</span>
    </Pill>
  );
}

export function ProgressBar({ value, indeterminate }: { value: number; indeterminate?: boolean }) {
  return (
    <div className="relative h-1 overflow-hidden rounded-full bg-surface-3">
      {indeterminate ? (
        <div className="progress-sweep absolute inset-y-0 w-1/3 rounded-full bg-volt" />
      ) : (
        <div className="h-full rounded-full bg-volt transition-[width] duration-500 ease-[var(--ease-out-quint)]" style={{ width: `${Math.max(2, Math.min(100, value * 100))}%` }} />
      )}
    </div>
  );
}

export function StatRow({ label, value, hint }: { label: string; value: ReactNode; hint?: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-2.5 shadow-[inset_0_-1px_0_var(--color-line)] last:shadow-none">
      <span className="text-body-regular text-ink-2">{label}</span>
      <span className="text-right">
        <span className="text-body-medium text-ink">{value}</span>
        {hint ? <span className="block text-caption-1-medium text-ink-3">{hint}</span> : null}
      </span>
    </div>
  );
}

export function Spark({ values, className }: { values: number[]; className?: string }) {
  if (values.length < 3) return null;
  const w = 220;
  const h = 36;
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = Math.max(4, max - min);
  const mid = (min + max) / 2;
  const lo = mid - span / 2;
  const pts = values.map((v, i) => `${(i / (values.length - 1)) * w},${h - 4 - ((v - lo) / span) * (h - 8)}`).join(" ");
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className={cx("h-9 w-[220px] overflow-visible", className)} aria-hidden>
      <polyline points={pts} fill="none" stroke="var(--color-volt)" strokeWidth="1.75" strokeLinejoin="round" strokeLinecap="round" vectorEffect="non-scaling-stroke" />
    </svg>
  );
}

export function Empty({ title, children, action }: { title: string; children?: ReactNode; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-2xl px-8 py-14 text-center shadow-[inset_0_0_0_1px_var(--color-line)]">
      <p className="text-[17px] font-semibold text-ink">{title}</p>
      {children ? <p className="mt-2 max-w-[44ch] text-body-regular text-ink-2">{children}</p> : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}
