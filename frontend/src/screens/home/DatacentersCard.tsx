import { useState } from "react";
import { Bar, BarChart, Cell, LabelList, ResponsiveContainer, XAxis, YAxis } from "recharts";
import { num } from "@/lib/format";
import type { Datacenter } from "@/lib/types";
import { Chip } from "@/components/ui/chip";
import { IconGlobe } from "@/components/icons";

export function DatacentersCard({ datacenters }: { datacenters: Datacenter[] }) {
  const [active, setActive] = useState<number | null>(null);
  const best = datacenters[0]?.rttMs ?? 0;
  const data = datacenters.slice(0, 6).map((d) => ({ ...d, label: d.ip.replace(/^128\.116\./, "…"), near: d.rttMs <= best + 8 }));
  const hovered = active !== null ? data[active] : null;
  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
          <IconGlobe className="size-4 text-ink-3" />
          Centros de datos
        </p>
        <Chip variant="caption" color="neutral">
          {datacenters.length} medidos
        </Chip>
      </div>
      <p className="mt-1 text-caption-1-medium text-ink-3">
        {hovered ? `${hovered.ip} · medido ${hovered.measured} ${hovered.measured === 1 ? "vez" : "veces"}` : "Ping real desde tu conexión. Jugar evita los grises."}
      </p>
      <div className="mt-2 min-h-[176px] flex-1">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} layout="vertical" margin={{ top: 0, right: 58, bottom: 0, left: 0 }} barCategoryGap={8} onMouseLeave={() => setActive(null)}>
            <XAxis type="number" hide domain={[0, Math.max(...data.map((d) => d.rttMs)) * 1.05]} />
            <YAxis type="category" dataKey="label" width={58} tickLine={false} axisLine={false} tick={{ fontSize: 11, fill: "var(--color-ink-3)" }} />
            <Bar dataKey="rttMs" radius={[0, 6, 6, 0]} maxBarSize={30} isAnimationActive animationDuration={500} onMouseEnter={(_, i) => setActive(i)}>
              {data.map((d, i) => (
                <Cell key={d.ip} fill={d.near ? "var(--color-volt)" : "var(--color-ink-4)"} fillOpacity={active === null || active === i ? 1 : 0.55} />
              ))}
              <LabelList dataKey="rttMs" position="right" formatter={(v: unknown) => `${num(Number(v), 0)} ms`} style={{ fontSize: 11, fill: "var(--color-ink-2)", fontWeight: 600 }} />
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </section>
  );
}
