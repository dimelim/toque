import { num } from "@/lib/format";
import { bestRegion, regionPlace } from "@/lib/regions";
import type { Region } from "@/lib/types";
import { cx } from "@/utils/cx";
import { Chip } from "@/components/ui/chip";
import { IconGlobe } from "@/components/icons";

export function DatacentersCard({ regions, max = 7 }: { regions: Region[]; max?: number }) {
  const best = bestRegion(regions)?.rttMs ?? 0;
  const measured = regions.filter((r) => r.rttMs > 0);
  const top = Math.max(1, ...measured.map((r) => r.rttMs));
  const rows = regions.slice(0, max);
  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
          <IconGlobe className="size-4 text-ink-3" />
          Centros de datos
        </p>
        <Chip variant="caption" color="neutral">
          {measured.length} medidos
        </Chip>
      </div>
      <p className="mt-1 text-caption-1-medium text-ink-3">Ping real desde tu conexión a cada ciudad. Jugar evita los grises.</p>
      {rows.length === 0 ? (
        <p className="flex flex-1 items-center justify-center py-8 text-center text-body-regular text-ink-3">Juega una partida con Toque y aquí aparecerán las ciudades de tus servidores.</p>
      ) : (
        <ul className="mt-3 flex flex-1 flex-col justify-between gap-2.5">
          {rows.map((r) => (
            <li key={r.key} title={regionPlace(r)}>
              <div className="flex items-baseline justify-between gap-3">
                <span className="truncate text-body-medium text-ink">{r.city}</span>
                {r.rttMs > 0 ? (
                  <span className="num shrink-0 text-[14px] font-semibold text-ink-2">{num(r.rttMs, 0)} ms</span>
                ) : (
                  <span className="shrink-0 text-caption-1-medium text-ink-4">sin medir</span>
                )}
              </div>
              <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-surface-3">
                {r.rttMs > 0 ? (
                  <div
                    className={cx("h-full rounded-full transition-[width] duration-500 ease-[var(--ease-out-quint)]", r.rttMs <= best + 8 ? "bg-volt" : "bg-ink-4")}
                    style={{ width: `${Math.max(4, (r.rttMs / top) * 100)}%` }}
                  />
                ) : null}
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
