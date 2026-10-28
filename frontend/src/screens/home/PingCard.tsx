import { useId, useMemo, useState } from "react";
import { Area, ComposedChart, Line, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { Chip } from "@/components/ui/chip";
import { useCountUp } from "@/hooks/use-count-up";
import { num } from "@/lib/format";
import { locale, t } from "@/lib/i18n";
import type { PingPoint } from "@/lib/types";
import { IconSignal } from "@/components/icons";

const clock = (ms: number) => new Date(ms).toLocaleTimeString(locale(), { hour: "2-digit", minute: "2-digit" });
const day = (ms: number) => new Date(ms).toLocaleDateString(locale(), { day: "numeric", month: "short" });

function ActiveDot({ cx: x, cy: y }: { cx?: number; cy?: number }) {
  if (x === undefined || y === undefined) return null;
  return (
    <g>
      <circle cx={x} cy={y} r={8} fill="var(--color-volt)" opacity={0.22} />
      <circle cx={x} cy={y} r={4} fill="var(--color-volt)" stroke="var(--color-surface-1)" strokeWidth={2} />
    </g>
  );
}

export function PingCard({ points, best }: { points: PingPoint[]; best: number | null }) {
  const [active, setActive] = useState<number | null>(null);
  const gradient = useId();
  const data = useMemo(() => points.map((p, i) => ({ i, t: new Date(p.at).getTime(), rtt: p.rttMs, server: p.serverIp })), [points]);
  const point = active !== null && active < data.length ? data[active] : data[data.length - 1];
  const value = useCountUp(Math.round(point?.rtt ?? 0));
  const gap = best !== null && point ? point.rtt - best : 0;
  const near = gap <= 8;
  const lo = Math.min(...data.map((d) => d.rtt), best ?? Infinity);
  const hi = Math.max(...data.map((d) => d.rtt));
  const step = hi - lo > 60 ? 20 : hi - lo > 25 ? 10 : 5;
  const yLo = Math.floor((lo - 2) / step) * step;
  const yHi = Math.ceil((hi + 2) / step) * step;
  const yTicks = Array.from({ length: Math.round((yHi - yLo) / step) + 1 }, (_, k) => yLo + k * step);
  const servers = new Set(points.map((p) => p.serverIp)).size;
  const summary = { n: num(data.length), servers, ms: num(best, 1) };
  const ticks = data.length > 1 ? [0, Math.floor((data.length - 1) / 2), data.length - 1] : [0];

  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default px-5 pt-5 pb-3 shadow-card">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
            <IconSignal className="size-4 text-ink-3" />
            {active !== null && point ? t("{day}, {clock} · servidor {server}", { day: day(point.t), clock: clock(point.t), server: point.server }) : t("Ping real dentro del juego")}
          </p>
          <div className="mt-1 flex items-center gap-2.5">
            <p key={active ?? "last"} className="animate-number-fade num text-[44px] leading-none font-semibold text-ink">
              {point ? value : "—"}
              <span className="ml-1 font-sans text-[15px] font-medium text-ink-3">ms</span>
            </p>
            {point && best !== null ? (
              <Chip variant="bold" color={near ? "lime" : "rose"}>
                {near ? t("Servidor cercano") : t("+{ms} ms sobre tu mejor", { ms: num(gap, 0) })}
              </Chip>
            ) : null}
          </div>
          <p className="mt-1.5 text-body-2-medium text-text-tertiary">
            {servers === 1 ? t("{n} mediciones · {servers} servidor · mejor centro {ms} ms", summary) : t("{n} mediciones · {servers} servidores · mejor centro {ms} ms", summary)}
          </p>
        </div>
        <dl className="flex shrink-0 flex-col items-end gap-1.5 text-body-2-medium text-text-secondary">
          <div className="flex items-center gap-1.5">
            <span className="h-0.5 w-3 rounded-full bg-volt" aria-hidden />
            <dt>{t("Ping medido")}</dt>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="h-0 w-3 border-t-2 border-dashed border-ink-4" aria-hidden />
            <dt>{t("Tu mejor centro")}</dt>
          </div>
        </dl>
      </div>
      <div className="mt-3 min-h-[150px] w-full flex-1">
        <ResponsiveContainer width="100%" height="100%">
          <ComposedChart
            data={data}
            margin={{ top: 6, right: 4, bottom: 0, left: 0 }}
            onMouseMove={(state) => {
              const index = Number(state?.activeTooltipIndex);
              if (state?.isTooltipActive && Number.isFinite(index)) setActive(index);
            }}
            onMouseLeave={() => setActive(null)}
          >
            <defs>
              <linearGradient id={gradient} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="var(--color-volt)" stopOpacity={0.28} />
                <stop offset="100%" stopColor="var(--color-volt)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <YAxis
              width={36}
              domain={[yLo, yHi]}
              ticks={yTicks}
              tickLine={false}
              axisLine={false}
              tick={{ fontSize: 11, fill: "var(--color-ink-4)" }}
            />
            <XAxis
              dataKey="i"
              type="number"
              domain={[0, Math.max(1, data.length - 1)]}
              ticks={ticks}
              tickFormatter={(i: number) => (data[i] ? clock(data[i].t) : "")}
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tick={{ fontSize: 11, fill: "var(--color-ink-4)" }}
            />
            <Tooltip content={() => null} cursor={{ stroke: "var(--color-chart-cursor)", strokeWidth: 1, strokeDasharray: "4 4" }} />
            {best !== null ? <ReferenceLine y={best} stroke="var(--color-ink-4)" strokeDasharray="5 5" strokeWidth={1.5} /> : null}
            <Area type="monotone" dataKey="rtt" stroke="none" fill={`url(#${gradient})`} isAnimationActive animationDuration={500} />
            <Line type="monotone" dataKey="rtt" stroke="var(--color-volt)" strokeWidth={2} dot={false} activeDot={<ActiveDot />} isAnimationActive animationDuration={500} />
          </ComposedChart>
        </ResponsiveContainer>
      </div>
    </section>
  );
}
