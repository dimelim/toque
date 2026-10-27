import { useState } from "react";
import type { ComponentType } from "react";
import { cx } from "@/utils/cx";
import { api } from "@/lib/api";
import { num, when } from "@/lib/format";
import { play } from "@/lib/sound";
import type { BloatResult, Diagnosis, Finding, PathStep } from "@/lib/types";
import { useApp } from "@/state";
import { Card, Empty, Eyebrow, PageHeader, Pill, StatRow } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { IconBad, IconDiagnose, IconGlobe, IconInfo, IconOk, IconPc, IconRouter, IconServer, IconWarn } from "@/components/icons";

const planNames: Record<string, string> = {
  "381b4222-f694-41f0-9685-ff5bb260df2e": "Equilibrado",
  "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c": "Alto rendimiento",
  "a1841308-3541-4fab-bc81-f71556f20b4a": "Economizador",
  "e9a42b02-d5df-448d-aa00-03f14749eb61": "Máximo rendimiento",
};

function Node({ icon: Icon, label, sub, value, tone }: { icon: ComponentType<{ className?: string }>; label: string; sub?: string; value?: number; tone?: "volt" | "ink" | "warn" }) {
  return (
    <div className="flex w-[150px] shrink-0 flex-col items-center text-center">
      <span className={cx("grid size-12 place-items-center rounded-2xl", tone === "volt" ? "bg-volt text-volt-ink" : "bg-surface-3 text-ink-2")}>
        <Icon className="size-5" />
      </span>
      <p className="mt-3 text-body-medium text-ink">{label}</p>
      {sub ? <p className="truncate text-caption-1-medium text-ink-3">{sub}</p> : null}
      {value !== undefined ? (
        <p className={cx("num mt-2 text-[30px] leading-none font-semibold", tone === "volt" ? "text-volt" : tone === "warn" ? "text-warn" : "text-ink")}>
          {num(value, value < 10 ? 1 : 0)}
          <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
        </p>
      ) : null}
    </div>
  );
}

function Link({ from, to }: { from: number; to: number }) {
  const added = Math.max(0, to - from);
  return (
    <div className="flex min-w-0 flex-1 flex-col items-center pt-6">
      <div className="relative h-px w-full bg-line-strong">
        <div className="absolute inset-y-0 left-0 bg-volt" style={{ width: `${Math.min(100, (added / Math.max(1, to)) * 100)}%` }} />
      </div>
      <p className="mt-2 text-caption-1-medium text-ink-3">+{num(added, added < 10 ? 1 : 0)} ms</p>
    </div>
  );
}

function PathView({ path }: { path: PathStep[] }) {
  const router = path.find((p) => p.kind === "local");
  const isp = [...path].reverse().find((p) => p.kind === "isp");
  const game = path.find((p) => p.kind === "game");
  const steps: { icon: ComponentType<{ className?: string }>; label: string; sub?: string; value?: number; tone?: "volt" | "ink" | "warn" }[] = [
    { icon: IconPc, label: "Tu PC" },
  ];
  if (router) steps.push({ icon: IconRouter, label: "Router", sub: router.addr, value: router.rttMs });
  if (isp) steps.push({ icon: IconGlobe, label: "Tu proveedor", sub: isp.addr, value: isp.rttMs });
  if (game) steps.push({ icon: IconServer, label: "Roblox", sub: game.addr, value: game.rttMs, tone: game.rttMs <= 80 ? "volt" : "warn" });
  return (
    <div className="flex items-start">
      {steps.map((s, i) => (
        <div key={s.label} className="contents">
          {i > 0 ? <Link from={steps[i - 1].value ?? 0} to={s.value ?? 0} /> : null}
          <Node {...s} />
        </div>
      ))}
    </div>
  );
}

const levelUI: Record<Finding["level"], { icon: ComponentType<{ className?: string }>; color: string; rank: number }> = {
  bad: { icon: IconBad, color: "text-bad", rank: 0 },
  warn: { icon: IconWarn, color: "text-warn", rank: 1 },
  info: { icon: IconInfo, color: "text-ink-2", rank: 2 },
  ok: { icon: IconOk, color: "text-volt", rank: 3 },
};

function FindingRow({ f }: { f: Finding }) {
  const { go } = useApp();
  const ui = levelUI[f.level];
  const Icon = ui.icon;
  return (
    <div className="flex items-start gap-4 px-5 py-4">
      <Icon className={cx("mt-0.5 size-[18px] shrink-0", ui.color)} />
      <div className="min-w-0 flex-1">
        <p className="text-body-medium text-ink">{f.title}</p>
        <p className="mt-0.5 max-w-[78ch] text-body-regular text-ink-2">{f.detail}</p>
      </div>
      {f.fix ? (
        <Button
          size="small"
          variant="secondary"
          onClick={() => {
            play("tap");
            go(f.fix === "server" ? "home" : "optimize");
          }}
        >
          {f.fix === "server" ? "Buscar servidor" : "Arreglar"}
        </Button>
      ) : null}
    </div>
  );
}

const gradeTone: Record<string, "volt" | "warn" | "bad" | "neutral"> = { "A+": "volt", A: "volt", B: "warn", C: "warn", D: "bad", F: "bad" };

function BloatCard() {
  const { run, task } = useApp();
  const [res, setRes] = useState<BloatResult | null>(null);
  const start = async () => {
    const r = await run("bloat", "Probando tu red bajo carga", () => api.bufferbloat());
    if (r) setRes(r);
  };
  const phases = res?.phases ?? [];
  const idle = phases[0]?.rttMs ?? 0;
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between gap-6">
        <div>
          <Eyebrow>Red bajo carga</Eyebrow>
          <p className="mt-3 text-body-medium text-ink">¿Se dispara el ping cuando alguien descarga?</p>
          <p className="mt-1 max-w-[60ch] text-body-regular text-ink-2">
            Satura tu conexión unos 40 segundos y mide el ping hasta tu proveedor en reposo, descargando y subiendo. Si sube mucho, los tirones de red vienen del router o del proveedor, no de tu PC.
          </p>
        </div>
        {res ? <Pill tone={gradeTone[res.grade] ?? "neutral"} className="h-8 px-3 text-[15px]">Nota {res.grade}</Pill> : null}
      </div>
      {phases.length > 0 ? (
        <div className="mt-5 grid grid-cols-3 gap-4">
          {phases.map((p) => (
            <div key={p.phase} className="rounded-xl bg-surface-2 p-4">
              <p className="text-caption-1-medium text-ink-3 capitalize">{p.phase}</p>
              <p className={cx("num mt-1 text-[30px] leading-none font-semibold", p.p95Ms - idle > 30 ? "text-warn" : "text-ink")}>
                {num(p.p95Ms, 0)}
                <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
              </p>
              <p className="mt-1 text-caption-1-medium text-ink-3">
                mediana {num(p.rttMs, 0)} ms · pérdida {num(p.lossPct, 0)} %{p.mbps > 0 ? ` · ${num(p.mbps, 0)} Mbps` : ""}
              </p>
            </div>
          ))}
        </div>
      ) : null}
      {res && (res.maxLossPct >= 2 || res.addedMs >= 30) ? (
        <div className="mt-4 flex items-start gap-3 rounded-xl bg-warn-soft p-4">
          <IconWarn className="mt-0.5 size-[18px] shrink-0 text-warn" />
          <p className="text-body-regular text-ink-2">
            {res.maxLossPct >= 2
              ? `Con la conexión saturada se pierde hasta el ${num(res.maxLossPct, 0)} % de los paquetes. En Realistic Street Soccer eso se nota como saltos del balón y de los jugadores. `
              : `Con la conexión saturada el ping sube ${num(res.addedMs, 0)} ms. `}
            Evita descargas y subidas grandes mientras juegas, o activa QoS o SQM en tu router limitándolo a un 90 % de tu velocidad.
          </p>
        </div>
      ) : null}
      <Button className="mt-5" size="small" variant="secondary" disabled={!!task} onClick={start}>
        {res ? "Repetir prueba" : "Probar mi red bajo carga"}
      </Button>
    </Card>
  );
}

export function Diagnose() {
  const { run, task } = useApp();
  const [withGame, setWithGame] = useState(true);
  const [d, setD] = useState<Diagnosis | null>(null);
  const start = async () => {
    const res = await run("diagnose", "Diagnosticando", () => api.diagnose(withGame));
    if (res) setD(res);
  };
  const findings = [...(d?.findings ?? [])].sort((a, b) => levelUI[a.level].rank - levelUI[b.level].rank);
  const problems = findings.filter((f) => f.level === "bad" || f.level === "warn").length;
  return (
    <div>
      <PageHeader eyebrow="Diagnóstico" title="Qué está frenando tu Roblox">
        <Switch size="sm" isSelected={withGame} onChange={setWithGame} isDisabled={!!task}>
          <span className="text-body-regular text-ink-2">Medir dentro del juego</span>
        </Switch>
        <Button leadingIcon={IconDiagnose} disabled={!!task} onClick={start}>
          Diagnosticar
        </Button>
      </PageHeader>

      {!d ? (
        <Empty title="Un diagnóstico tarda alrededor de un minuto">
          Toque mide la ruta de red hasta los servidores de Roblox, revisa la memoria, el disco y el plan de energía y, si quieres, entra en Realistic Street Soccer para medir tus FPS y tu ping real.
        </Empty>
      ) : (
        <div className="flex flex-col gap-4">
          <Card className="p-6">
            <div className="flex items-baseline justify-between">
              <Eyebrow>Tu conexión, tramo a tramo</Eyebrow>
              <span className="text-caption-1-medium text-ink-3">{when(d.at)}</span>
            </div>
            <div className="mt-6">
              {d.path && d.path.length > 0 ? <PathView path={d.path} /> : <p className="text-body-regular text-ink-2">No pude medir la ruta de red.</p>}
            </div>
            {!d.path?.some((p) => p.kind === "game") ? (
              <p className="mt-5 text-caption-1-medium text-ink-3">Activa "Medir dentro del juego" para añadir el ping real hasta el servidor de Roblox.</p>
            ) : null}
          </Card>

          <Card className="overflow-hidden">
            <div className="flex items-center justify-between px-5 pt-5 pb-2">
              <Eyebrow>Hallazgos</Eyebrow>
              <Pill tone={problems > 0 ? "warn" : "volt"}>{problems > 0 ? `${problems} por revisar` : "Todo bien"}</Pill>
            </div>
            <div className="divide-y divide-line">
              {findings.map((f) => (
                <FindingRow key={f.id} f={f} />
              ))}
            </div>
          </Card>

          <div className="grid grid-cols-2 gap-4">
            <Card className="p-5">
              <Eyebrow>Tu equipo</Eyebrow>
              <div className="mt-3">
                <StatRow label="Procesador" value={d.system.cpu || "—"} />
                <StatRow label="Gráfica" value={d.system.gpu || "—"} />
                <StatRow label="Memoria" value={`${num(d.system.ramGb, 0)} GB`} hint={`${num(d.ramAvailMb, 0)} MB libres ahora`} />
                <StatRow label="Monitor" value={`${d.system.refreshHz} Hz`} hint={d.framerateCap > 0 ? `Roblox limitado a ${d.framerateCap} FPS` : undefined} />
                <StatRow label="Paginación" value={d.pagefile.drives?.length ? `${d.pagefile.drives.join(", ")}: ${d.pagefile.onHdd ? "disco mecánico" : "SSD"}` : "—"} />
                <StatRow label="Plan de energía" value={d.powerName || planNames[d.powerPlan] || "Personalizado"} />
              </div>
            </Card>
            <Card className="p-5">
              <Eyebrow>Lo que más memoria usa</Eyebrow>
              <div className="mt-3">
                {(d.hogs ?? []).map((h) => (
                  <StatRow key={h.name} label={h.name} value={`${num(h.mb, 0)} MB`} hint={h.count > 1 ? `${h.count} procesos` : undefined} />
                ))}
              </div>
              {d.game ? (
                <div className="mt-4 grid grid-cols-3 gap-4 rounded-xl bg-surface-2 p-4">
                  <div>
                    <p className="text-caption-1-medium text-ink-3">FPS medio</p>
                    <p className="num text-[26px] font-semibold">{num(d.game.frames.avgFps)}</p>
                  </div>
                  <div>
                    <p className="text-caption-1-medium text-ink-3">1 % low</p>
                    <p className="num text-[26px] font-semibold">{num(d.game.frames.low1Fps)}</p>
                  </div>
                  <div>
                    <p className="text-caption-1-medium text-ink-3">Tirones/min</p>
                    <p className="num text-[26px] font-semibold">{num(d.game.frames.stutterPerMin)}</p>
                  </div>
                </div>
              ) : null}
            </Card>
          </div>
        </div>
      )}
      <div className="mt-4">
        <BloatCard />
      </div>
    </div>
  );
}
