import { useState } from "react";
import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { bestRegion, regionCity } from "@/lib/regions";
import type { Region, ServerProbe } from "@/lib/types";
import { usePlay } from "@/hooks/use-play";
import { useApp } from "@/state";
import { cx } from "@/utils/cx";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { StatusDot } from "@/components/ui/status-dot";
import { RegionSelect } from "@/components/RegionSelect";
import { IconPlay, IconSignal } from "@/components/icons";

function Stat({ label, value, tone }: { label: string; value: number | null | undefined; tone: "volt" | "ink" }) {
  return (
    <div className="rounded-lg px-3 py-2">
      <p className="text-caption-1-medium text-ink-3">{label}</p>
      <p className={cx("num mt-0.5 text-[26px] leading-none font-semibold", tone === "volt" ? "text-volt" : "text-ink")}>
        {num(value, 0)}
        <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
      </p>
    </div>
  );
}

export function PlayCard({ regions, onDone }: { regions: Region[]; onDone: () => void }) {
  const { overview, run, task, region, setRegion } = useApp();
  const start = usePlay();
  const [probe, setProbe] = useState<ServerProbe | null>(null);
  const r = overview?.roblox;
  const best = bestRegion(regions);
  const fresh = probe && r?.inGame && probe.jobId === r.jobId ? probe : null;
  const status = !r?.installed ? "Roblox no está instalado" : r.inGame ? "En partida" : r.running ? "Roblox abierto" : "Roblox cerrado";
  const city = region ? regionCity(region) : "";

  const measure = async () => {
    const p = await run("ping", "Midiendo tu ping ahora", () => api.quickPing());
    if (p) {
      setProbe(p);
      onDone();
    }
  };

  return (
    <section className="edge-light flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <span className="flex items-center gap-2 text-body-medium text-text-secondary">
          <StatusDot color={r?.inGame ? "green" : r?.running ? "yellow" : "gray"} />
          {status}
        </span>
        {r?.inGame ? <Chip variant="caption" color="neutral">{r.serverIp}</Chip> : <Chip variant="caption" color="lime">Realistic Street Soccer</Chip>}
      </div>
      <h2 className="mt-4 text-title-1-medium text-ink">{r?.inGame ? "Ya estás jugando" : city ? `Jugar en ${city}` : "Jugar en un servidor cercano"}</h2>
      <p className="mt-1.5 text-body-regular text-text-secondary">
        {r?.inGame
          ? "Mide el ping del servidor en el que estás, escuchando el tráfico real del juego durante 8 segundos."
          : city
            ? `Toque prueba servidores públicos y descarta los que no son de ${city} hasta dar con uno. Suele tardar entre medio minuto y dos.`
            : "Roblox te manda a un servidor lejano en una de cada seis partidas. Toque lo comprueba al entrar y, si pasa, te mueve a uno cercano."}
      </p>
      <div className="mt-4">
        {r?.inGame ? (
          <div className="grid grid-cols-2 gap-2 rounded-2lg bg-background-inner-default p-1.5 shadow-card">
            <Stat label="Ping ahora" value={fresh?.rttMs} tone="volt" />
            <Stat label="Tu mejor centro" value={best?.rttMs} tone="ink" />
          </div>
        ) : (
          <RegionSelect regions={regions} value={region} onChange={setRegion} isDisabled={!!task} />
        )}
      </div>
      <div className="mt-auto pt-4">
        {r?.inGame ? (
          <Button className="h-11 w-full text-[15px]" leadingIcon={IconSignal} disabled={!!task} onClick={measure}>
            {fresh ? "Medir otra vez" : "Medir mi ping ahora"}
          </Button>
        ) : (
          <Button
            className="h-11 w-full text-[15px]"
            leadingIcon={IconPlay}
            disabled={!!task || !r?.installed || r.running}
            onClick={async () => {
              if (await start()) onDone();
            }}
          >
            Jugar
          </Button>
        )}
      </div>
    </section>
  );
}
