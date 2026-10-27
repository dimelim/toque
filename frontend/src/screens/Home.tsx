import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { num } from "@/lib/format";
import type { Datacenter, Optimization, PingPoint, RunRecord } from "@/lib/types";
import { useApp } from "@/state";
import { Button } from "@/components/ui/button";
import { IconStopwatch } from "@/components/icons";
import { PingCard } from "./home/PingCard";
import { PlayCard } from "./home/PlayCard";
import { StatTiles } from "./home/StatTiles";
import { OptimizationsCard } from "./home/OptimizationsCard";
import { DatacentersCard } from "./home/DatacentersCard";
import { SystemCard } from "./home/SystemCard";

export function Home() {
  const { overview, run, task, toast, refresh } = useApp();
  const [points, setPoints] = useState<PingPoint[]>([]);
  const [dcs, setDcs] = useState<Datacenter[]>([]);
  const [opts, setOpts] = useState<Optimization[]>([]);
  const [runs, setRuns] = useState<RunRecord[]>([]);
  const load = useCallback(async () => {
    try {
      const [p, d, o, r] = await Promise.all([api.pingHistory(), api.datacenters(), api.optimizations(), api.runs()]);
      setPoints(p ?? []);
      setDcs(d ?? []);
      setOpts(o ?? []);
      setRuns(r ?? []);
    } catch {
      return;
    }
  }, []);
  useEffect(() => {
    load();
  }, [load]);
  const reload = () => {
    load();
    refresh();
  };
  const measure = async () => {
    const rec = await run("benchmark", "Midiendo una partida", () => api.benchmark(""));
    if (rec) {
      toast({ tone: "ok", title: "Partida medida", detail: `${num(rec.window.frames.avgFps)} FPS y ${num(rec.window.net.rttMs)} ms de ping` });
      reload();
    }
  };
  const proven = opts.filter((o) => o.verdict === "proven" || o.verdict === "tradeoff");
  const best = dcs[0]?.rttMs ?? null;
  return (
    <div className="flex flex-col gap-4 pb-2">
      <header className="flex items-end justify-between gap-6">
        <div className="min-w-0">
          <h1 className="text-[26px] leading-tight font-semibold tracking-[-0.02em] text-ink">Realistic Street Soccer</h1>
          <p className="mt-0.5 text-body-regular text-ink-3">
            {overview?.lastRun ? `Última partida medida con ${num(overview.lastRun.window.frames.avgFps)} FPS en el servidor ${overview.lastRun.serverIp}` : "Mide una partida para ver cómo va tu Roblox"}
          </p>
        </div>
        <Button variant="secondary" leadingIcon={IconStopwatch} disabled={!!task} onClick={measure}>
          Medir una partida
        </Button>
      </header>

      <div className="grid grid-cols-[minmax(0,1.7fr)_minmax(0,1fr)] gap-4">
        <PingCard points={points} best={best} />
        <PlayCard datacenters={dcs} onDone={reload} />
      </div>

      <StatTiles runs={runs} />

      <div className="grid grid-cols-[minmax(0,1.15fr)_minmax(0,1fr)_minmax(0,1fr)] gap-4">
        <OptimizationsCard items={proven} all={opts} onChange={reload} />
        <DatacentersCard datacenters={dcs} />
        <SystemCard />
      </div>
    </div>
  );
}
