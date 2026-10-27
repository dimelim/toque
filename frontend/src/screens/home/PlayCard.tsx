import { useState } from "react";
import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { play } from "@/lib/sound";
import type { Datacenter, PlayResult, ServerProbe } from "@/lib/types";
import { useApp } from "@/state";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { StatusDot } from "@/components/ui/status-dot";
import { IconPlay, IconSignal } from "@/components/icons";

function message(r: PlayResult) {
  if (!r.final.rttMs) return { title: "Dentro de Realistic Street Soccer" };
  if (r.final.jobId !== r.first.jobId) return { title: `Te moví a un servidor a ${num(r.final.rttMs, 0)} ms`, detail: `Roblox te había puesto en uno a ${num(r.first.rttMs, 0)} ms.` };
  if (r.hops > 0) return { title: `Servidor a ${num(r.final.rttMs, 0)} ms`, detail: `Probé ${r.hops} más y ninguno estaba más cerca.` };
  return { title: `Servidor cercano, a ${num(r.final.rttMs, 0)} ms` };
}

export function PlayCard({ datacenters, onDone }: { datacenters: Datacenter[]; onDone: () => void }) {
  const { overview, run, task, toast } = useApp();
  const [probe, setProbe] = useState<ServerProbe | null>(null);
  const r = overview?.roblox;
  const best = datacenters[0]?.rttMs ?? null;
  const near = datacenters.filter((d) => best !== null && d.rttMs <= best + 8).length;
  const far = datacenters.length - near;
  const fresh = probe && r?.inGame && probe.jobId === r.jobId ? probe : null;
  const status = !r?.installed ? "Roblox no está instalado" : r.inGame ? "En partida" : r.running ? "Roblox abierto" : "Roblox cerrado";

  const start = async () => {
    play("tap");
    const res = await run("play", "Entrando en un servidor cercano", () => api.play());
    if (res) {
      toast({ tone: "ok", ...message(res) });
      onDone();
    }
  };
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
      <h2 className="mt-4 text-title-1-medium text-ink">{r?.inGame ? "Ya estás jugando" : "Jugar en un servidor cercano"}</h2>
      <p className="mt-1.5 text-body-regular text-text-secondary">
        {r?.inGame
          ? "Mide el ping del servidor en el que estás, escuchando el tráfico real del juego durante 8 segundos."
          : "Roblox te manda a un servidor lejano en una de cada seis partidas. Toque lo comprueba al entrar y, si pasa, te mueve a uno cercano."}
      </p>
      <div className="mt-4 grid grid-cols-2 gap-2 rounded-2lg bg-background-inner-default p-1.5 shadow-card">
        <div className="rounded-lg px-3 py-2">
          <p className="text-caption-1-medium text-ink-3">{fresh ? "Ping ahora" : "Tu mejor centro"}</p>
          <p className="num mt-0.5 text-[26px] leading-none font-semibold text-volt">
            {num(fresh ? fresh.rttMs : best, 0)}
            <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
          </p>
        </div>
        <div className="rounded-lg px-3 py-2">
          <p className="text-caption-1-medium text-ink-3">Centros evitados</p>
          <p className="num mt-0.5 text-[26px] leading-none font-semibold text-ink">
            {num(far)}
            <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">de {num(datacenters.length)}</span>
          </p>
        </div>
      </div>
      <div className="mt-auto pt-4">
        {r?.inGame ? (
          <Button className="h-11 w-full text-[15px]" leadingIcon={IconSignal} disabled={!!task} onClick={measure}>
            {fresh ? "Medir otra vez" : "Medir mi ping ahora"}
          </Button>
        ) : (
          <Button className="h-11 w-full text-[15px]" leadingIcon={IconPlay} disabled={!!task || !r?.installed || r.running} onClick={start}>
            Jugar
          </Button>
        )}
      </div>
    </section>
  );
}
