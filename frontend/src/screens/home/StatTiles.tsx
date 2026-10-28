import type { ComponentType } from "react";
import { useCountUp } from "@/hooks/use-count-up";
import { improvement, num, pct, shortDay } from "@/lib/format";
import { t } from "@/lib/i18n";
import type { RunRecord } from "@/lib/types";
import { Chip } from "@/components/ui/chip";
import { IconTile } from "@/components/IconTile";
import type { TileTone } from "@/components/IconTile";
import { IconChartLine, IconSpeed, IconStopwatch, IconStutter } from "@/components/icons";

interface Stat {
  key: string;
  label: string;
  unit: string;
  digits: number;
  tone: TileTone;
  icon: ComponentType<{ className?: string }>;
  hint: string;
  get: (r: RunRecord) => number;
}

const stats = (): Stat[] => [
  { key: "fps", label: t("FPS medio"), unit: "FPS", digits: 0, tone: "lime", icon: IconSpeed, hint: t("frames por segundo"), get: (r) => r.window.frames.avgFps },
  { key: "low1", label: t("1 % low"), unit: "FPS", digits: 0, tone: "cyan", icon: IconChartLine, hint: t("los peores momentos"), get: (r) => r.window.frames.low1Fps },
  { key: "latMs", label: t("Latencia de render"), unit: "ms", digits: 1, tone: "amber", icon: IconStopwatch, hint: t("de la CPU a la pantalla"), get: (r) => r.window.frames.renderLatencyMs },
  { key: "stutterMin", label: t("Tirones"), unit: t("/min"), digits: 0, tone: "rose", icon: IconStutter, hint: t("frames que se atascan"), get: (r) => r.window.frames.stutterPerMin },
];

function Tile({ stat, last, prev }: { stat: Stat; last: RunRecord | null; prev: RunRecord | null }) {
  const value = last ? stat.get(last) : null;
  const scale = stat.digits ? 10 : 1;
  const shown = useCountUp(Math.round((value ?? 0) * scale)) / scale;
  const before = prev ? stat.get(prev) : null;
  const d = value !== null && before ? (100 * (value - before)) / Math.abs(before) : null;
  const good = d !== null ? improvement(stat.key, d) : 0;
  const Icon = stat.icon;
  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-2 shadow-card">
      <div className="flex items-center gap-3 p-2">
        <IconTile icon={Icon} tone={stat.tone} />
        <div className="min-w-0">
          <p className="truncate text-body-medium text-text-primary">{stat.label}</p>
          <p className="truncate text-caption-1-medium text-ink-3">{stat.hint}</p>
        </div>
      </div>
      <p className="num px-2 pt-2 pb-3 text-[38px] leading-none font-semibold text-ink">
        {value === null ? "—" : num(shown, stat.digits)}
        <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">{stat.unit}</span>
      </p>
      <div className="mt-auto flex h-9 items-center justify-between gap-2 rounded-2lg bg-background-inner-default pe-1.5 ps-2.5 shadow-card">
        <p className="truncate text-caption-1-medium text-text-secondary">{prev ? t("Frente a la anterior") : last ? t("Medido el {day}", { day: shortDay(last.at) }) : t("Sin medir todavía")}</p>
        {d !== null ? (
          <Chip variant="caption" color={Math.abs(d) < 2 ? "neutral" : good > 0 ? "lime" : "rose"}>
            {pct(d, Math.abs(d) < 10 ? 1 : 0)}
          </Chip>
        ) : null}
      </div>
    </section>
  );
}

export function StatTiles({ runs }: { runs: RunRecord[] }) {
  const last = runs[0] ?? null;
  const prev = runs[1] ?? null;
  return (
    <div className="grid grid-cols-4 gap-4">
      {stats().map((s) => (
        <Tile key={s.key} stat={s} last={last} prev={prev} />
      ))}
    </div>
  );
}
