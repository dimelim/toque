import { useState } from "react";
import { cx } from "@/utils/cx";
import { api } from "@/lib/api";
import { num, when } from "@/lib/format";
import { play } from "@/lib/sound";
import type { ServerProbe, ServerSearch } from "@/lib/types";
import { useApp } from "@/state";
import { BigNumber, Card, Eyebrow, PageHeader, Pill, Spark } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { IconArrowRight, IconDiagnose, IconOptimize, IconSearch, IconStopwatch } from "@/components/icons";

function pingTone(ms: number | undefined) {
  if (!ms) return "muted" as const;
  return ms <= 80 ? ("volt" as const) : ("ink" as const);
}

function ServerFinder() {
  const { run, task, toast } = useApp();
  const [result, setResult] = useState<ServerSearch | null>(null);
  const start = async () => {
    const res = await run("finder", "Buscando el servidor con menos ping", () => api.findServer(3));
    if (res) {
      setResult(res);
      if (res.best?.current) {
        toast({ tone: "ok", title: `Tu servidor ya era el mejor (${num(res.best.rttMs, 0)} ms)`, detail: "Probé otros y todos tenían más ping. Te dejé donde estabas." });
      } else if (res.best) {
        toast({ tone: "ok", title: `Mejor servidor a ${num(res.best.rttMs, 0)} ms`, detail: res.joined ? "Ya estás dentro." : "No pude entrar, pruébalo de nuevo." });
      }
    }
  };
  const probes = (result?.probes ?? []).filter((p) => p.rttMs > 0).sort((a, b) => a.rttMs - b.rttMs);
  const worst = Math.max(1, ...probes.map((p) => p.rttMs));
  return (
    <Card className="flex flex-col p-5">
      <div className="flex items-center gap-2 text-ink-2">
        <IconSearch className="size-4" />
        <Eyebrow>Servidor</Eyebrow>
      </div>
      <h3 className="mt-3 text-[17px] font-semibold text-ink">El servidor con menos ping</h3>
      {probes.length === 0 ? (
        <p className="mt-1.5 text-body-regular text-ink-2">Al pulsar Jugar, Toque ya evita los servidores lejanos. Aquí entra en varios, mide el ping real de cada uno y te deja en el mejor.</p>
      ) : (
        <ul className="mt-3 flex flex-col gap-2">
          {probes.slice(0, 4).map((p) => {
            const best = result?.best?.jobId === p.jobId;
            return (
              <li key={p.jobId} className="flex items-center gap-3">
                <span className={cx("num w-14 text-right text-[18px] font-semibold", best ? "text-volt" : "text-ink-2")}>{num(p.rttMs, 0)}</span>
                <span className="h-1.5 flex-1 overflow-hidden rounded-full bg-surface-3">
                  <span className={cx("block h-full rounded-full", best ? "bg-volt" : "bg-before")} style={{ width: `${(p.rttMs / worst) * 100}%` }} />
                </span>
                <span className="w-20 text-caption-1-medium text-ink-3">{p.current ? "el tuyo" : best ? "el mejor" : `${p.playing}/${p.max}`}</span>
              </li>
            );
          })}
        </ul>
      )}
      <div className="mt-auto pt-5">
        <Button variant="secondary" size="small" leadingIcon={IconSearch} disabled={!!task} onClick={start}>
          Buscar servidor
        </Button>
      </div>
    </Card>
  );
}

function LivePing() {
  const { overview, run, task } = useApp();
  const [probe, setProbe] = useState<ServerProbe | null>(null);
  if (!overview?.roblox.inGame) return null;
  const measure = async () => {
    const p = await run("ping", "Midiendo tu ping ahora", () => api.quickPing());
    if (p) setProbe(p);
  };
  const fresh = probe && probe.jobId === overview.roblox.jobId ? probe : null;
  return (
    <Card className="mb-4 flex items-center gap-5 px-5 py-4">
      <span className="live-dot relative size-2.5 rounded-full bg-volt text-volt" />
      <div className="min-w-0 flex-1">
        <p className="text-body-medium text-ink">Estás en partida</p>
        <p className="truncate text-caption-1-medium text-ink-3">Servidor {overview.roblox.serverIp}</p>
      </div>
      {fresh ? (
        <p className={cx("num text-[30px] leading-none font-semibold", fresh.rttMs <= 80 ? "text-volt" : "text-warn")}>
          {num(fresh.rttMs, 0)}
          <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
        </p>
      ) : null}
      <Button size="small" variant="secondary" disabled={!!task} onClick={measure}>
        {fresh ? "Medir de nuevo" : "Medir mi ping ahora"}
      </Button>
    </Card>
  );
}

export function Home() {
  const { overview, go, run, task, refresh } = useApp();
  const last = overview?.lastRun;
  const w = last?.window;
  const inGame = overview?.roblox.inGame;
  const benchmark = async () => {
    const rec = await run("benchmark", "Midiendo una partida", () => api.benchmark(""));
    if (rec) refresh();
  };
  return (
    <div>
      <PageHeader eyebrow="Tu partida" title="Realistic Street Soccer">
        <Button variant="secondary" leadingIcon={IconDiagnose} onClick={() => { play("tap"); go("diagnose"); }}>
          Diagnosticar
        </Button>
      </PageHeader>

      <LivePing />
      <Card className="pitch-lines relative overflow-hidden p-7">
        {w ? (
          <div className="grid grid-cols-[1.1fr_1fr] gap-8">
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                <Pill tone={inGame ? "volt" : "neutral"}>{inGame ? "En partida ahora" : "Última medición"}</Pill>
                <span className="text-caption-1-medium text-ink-3">{when(last!.at)}</span>
              </div>
              <div className="mt-6">
                <BigNumber value={w.net.confident ? w.net.rttMs : null} unit="ms" label="Ping real dentro del juego" size="xl" digits={0} tone={pingTone(w.net.rttMs)} />
                {w.net.rttSeries && w.net.rttSeries.length >= 3 ? <Spark values={w.net.rttSeries} className="mt-3" /> : null}
              </div>
              <p className="mt-3 max-w-[38ch] text-body-regular text-ink-2">
                Medido sobre el tráfico real de Roblox, sin tocar el juego. Variación de {num(w.net.jitterMs, 1)} ms y {num(w.net.lossDownPct, 1)} % de pérdida.
              </p>
              <div className="mt-auto flex items-center gap-2 pt-6">
                <Button size="small" variant="secondary" leadingIcon={IconStopwatch} disabled={!!task} onClick={benchmark}>
                  Nueva prueba
                </Button>
                <span className="text-caption-1-medium text-ink-3">Servidor {last!.serverIp}</span>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-x-6 gap-y-7 self-center">
              <BigNumber value={w.frames.avgFps} label="FPS medio" />
              <BigNumber value={w.frames.low1Fps} label="1 % low" />
              <BigNumber value={w.frames.stutterPerMin} label="Tirones por minuto" />
              <BigNumber value={w.frames.renderLatencyMs} unit="ms" label="Latencia de render" digits={1} />
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-[1.1fr_1fr] gap-8">
            <div className="flex flex-col items-start py-4">
              <Eyebrow>Primer paso</Eyebrow>
              <h2 className="mt-3 max-w-[22ch] text-[30px] leading-[1.1] font-semibold tracking-[-0.02em]">Mide cómo va tu Roblox hoy</h2>
              <p className="mt-3 max-w-[48ch] text-body-regular text-ink-2">
                Toque abre Realistic Street Soccer, mide tu ping real, tus FPS y los tirones durante 30 segundos y lo guarda como punto de partida. Así cada cambio se compara contra algo medido.
              </p>
              <Button className="mt-6" leadingIcon={IconStopwatch} disabled={!!task} onClick={benchmark}>
                Hacer la primera prueba
              </Button>
            </div>
            <div className="grid grid-cols-2 gap-x-6 gap-y-7 self-center opacity-40" aria-hidden>
              <BigNumber value={null} unit="ms" label="Ping real" />
              <BigNumber value={null} label="FPS medio" />
              <BigNumber value={null} label="1 % low" />
              <BigNumber value={null} label="Tirones por minuto" />
            </div>
          </div>
        )}
      </Card>

      <div className="mt-4 grid grid-cols-3 gap-4">
        <Card className="flex flex-col p-5">
          <div className="flex items-center gap-2 text-ink-2">
            <IconOptimize className="size-4" />
            <Eyebrow>Optimizar</Eyebrow>
          </div>
          <h3 className="mt-3 text-[17px] font-semibold text-ink">Solo lo que demostró funcionar</h3>
          <p className="mt-1.5 text-body-regular text-ink-2">
            {!overview
              ? "Cargando"
              : overview.proven > 0
                ? `${overview.proven} de ${overview.tested} cambios típicos mejoraron de forma medible. ${overview.applied} aplicados en este PC.`
                : `Probamos ${overview.tested} cambios típicos con mediciones reales. Mira cuáles ayudan según tu PC y cuáles no sirven.`}
          </p>
          <div className="mt-auto pt-5">
            <Button variant="secondary" size="small" trailingIcon={IconArrowRight} onClick={() => { play("tap"); go("optimize"); }}>
              Ver optimizaciones
            </Button>
          </div>
        </Card>
        <ServerFinder />
        <Card className="flex flex-col p-5">
          <div className="flex items-center gap-2 text-ink-2">
            <IconDiagnose className="size-4" />
            <Eyebrow>Diagnóstico</Eyebrow>
          </div>
          <h3 className="mt-3 text-[17px] font-semibold text-ink">Qué te está frenando</h3>
          <p className="mt-1.5 text-body-regular text-ink-2">Revisa tu red tramo a tramo, la memoria, el disco y cómo dibuja Roblox en tu pantalla.</p>
          <div className="mt-auto pt-5">
            <Button variant="secondary" size="small" trailingIcon={IconArrowRight} onClick={() => { play("tap"); go("diagnose"); }}>
              Abrir diagnóstico
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
}
