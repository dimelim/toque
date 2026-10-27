import { useCallback, useEffect, useMemo, useState } from "react";
import { cx } from "@/utils/cx";
import { api, errorText } from "@/lib/api";
import { improvement, metricMeta, num, pct, when } from "@/lib/format";
import { play } from "@/lib/sound";
import type { ExperimentRecord, Optimization, RunRecord } from "@/lib/types";
import { useApp } from "@/state";
import { Card, Empty, Eyebrow, PageHeader, Pill } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { IconStopwatch, IconTrash } from "@/components/icons";

type Row = { key: string; get: (r: RunRecord) => number | null };

const rows: Row[] = [
  { key: "rtt", get: (r) => (r.window.net.confident ? r.window.net.rttMs : null) },
  { key: "jitter", get: (r) => (r.window.net.confident ? r.window.net.jitterMs : null) },
  { key: "fps", get: (r) => r.window.frames.avgFps },
  { key: "low1", get: (r) => r.window.frames.low1Fps },
  { key: "low01", get: (r) => r.window.frames.low01Fps },
  { key: "stutterMin", get: (r) => r.window.frames.stutterPerMin },
  { key: "latMs", get: (r) => r.window.frames.renderLatencyMs },
  { key: "hardFaults", get: (r) => r.window.sys.hardFaultsPerSec },
];

function RunPicker({ runs, value, onChange, label }: { runs: RunRecord[]; value: string; onChange: (id: string) => void; label: string }) {
  return (
    <label className="flex min-w-0 flex-col gap-1.5">
      <span className="text-caption-1-medium text-ink-3">{label}</span>
      <select
        value={value}
        onChange={(e) => {
          play("tap");
          onChange(e.target.value);
        }}
        className="h-9 min-w-0 rounded-[10px] bg-surface-2 px-3 text-body-medium text-ink shadow-[inset_0_0_0_1px_var(--color-line-strong)] outline-none focus-visible:ring-2 focus-visible:ring-border-focus-ring"
      >
        {runs.map((r) => (
          <option key={r.id} value={r.id}>
            {r.label} · {when(r.at)}
          </option>
        ))}
      </select>
    </label>
  );
}

function Compare({ runs }: { runs: RunRecord[] }) {
  const [a, setA] = useState(runs[runs.length > 1 ? 1 : 0]?.id ?? "");
  const [b, setB] = useState(runs[0]?.id ?? "");
  useEffect(() => {
    setB(runs[0]?.id ?? "");
    setA(runs[runs.length > 1 ? 1 : 0]?.id ?? "");
  }, [runs]);
  const ra = runs.find((r) => r.id === a);
  const rb = runs.find((r) => r.id === b);
  if (!ra || !rb) return null;
  return (
    <Card className="p-6">
      <div className="grid grid-cols-2 gap-4">
        <RunPicker runs={runs} value={a} onChange={setA} label="Antes" />
        <RunPicker runs={runs} value={b} onChange={setB} label="Después" />
      </div>
      <div className="mt-6 grid grid-cols-[1fr_auto_auto_auto] items-center gap-x-6">
        <span />
        <span className="label-caps w-24 text-right text-ink-3">Antes</span>
        <span className="label-caps w-24 text-right text-volt">Después</span>
        <span className="label-caps w-24 text-right text-ink-3">Cambio</span>
        {rows.map((row) => {
          const meta = metricMeta[row.key];
          const va = row.get(ra);
          const vb = row.get(rb);
          const d = va && vb ? (100 * (vb - va)) / Math.abs(va) : null;
          const good = d !== null ? improvement(row.key, d) : 0;
          return (
            <div key={row.key} className="contents">
              <span className="py-3 text-body-regular text-ink-2 shadow-[inset_0_-1px_0_var(--color-line)]">{meta.label}</span>
              <span className="num py-3 text-right text-[22px] font-semibold text-before shadow-[inset_0_-1px_0_var(--color-line)]">
                {num(va, meta.digits)}
                <span className="ml-1 font-sans text-[11px] font-medium text-ink-4">{meta.unit}</span>
              </span>
              <span className="num py-3 text-right text-[22px] font-semibold text-ink shadow-[inset_0_-1px_0_var(--color-line)]">
                {num(vb, meta.digits)}
                <span className="ml-1 font-sans text-[11px] font-medium text-ink-4">{meta.unit}</span>
              </span>
              <span
                className={cx(
                  "num py-3 text-right text-[18px] font-semibold shadow-[inset_0_-1px_0_var(--color-line)]",
                  d === null || Math.abs(d) < 2 ? "text-ink-3" : good > 0 ? "text-volt" : "text-bad",
                )}
              >
                {d === null ? "—" : pct(d, Math.abs(d) < 10 ? 1 : 0)}
              </span>
            </div>
          );
        })}
      </div>
      <p className="mt-4 text-caption-1-medium text-ink-3">
        Una sola prueba puede variar por el servidor o los jugadores en pantalla. Para afirmar que algo mejora, usa "Probar en mi PC" en cada optimización, que alterna con y sin el cambio varias veces.
      </p>
    </Card>
  );
}

const verdictUI: Record<string, { label: string; tone: "volt" | "neutral" | "bad" | "warn" }> = {
  improves: { label: "Mejora", tone: "volt" },
  "no-effect": { label: "Sin efecto", tone: "neutral" },
  worse: { label: "Empeora", tone: "bad" },
  tradeoff: { label: "Compromiso", tone: "warn" },
};

export function Tests() {
  const { run, task, toast } = useApp();
  const [runs, setRuns] = useState<RunRecord[]>([]);
  const [exps, setExps] = useState<ExperimentRecord[]>([]);
  const [opts, setOpts] = useState<Optimization[]>([]);
  const load = useCallback(async () => {
    try {
      const [r, e, o] = await Promise.all([api.runs(), api.experiments(), api.optimizations()]);
      setRuns(r ?? []);
      setExps(e ?? []);
      setOpts(o ?? []);
    } catch (e) {
      toast({ tone: "error", title: "No pude leer el historial", detail: errorText(e) });
    }
  }, [toast]);
  useEffect(() => {
    load();
  }, [load]);
  const titles = useMemo(() => Object.fromEntries(opts.map((o) => [o.id, o.title])), [opts]);
  const newRun = async () => {
    const rec = await run("benchmark", "Midiendo una partida", () => api.benchmark(""));
    if (rec) {
      toast({ tone: "ok", title: "Prueba guardada", detail: `${num(rec.window.frames.avgFps)} FPS, ${num(rec.window.net.rttMs)} ms de ping` });
      load();
    }
  };
  return (
    <div>
      <PageHeader eyebrow="Pruebas" title="Antes y después, medido">
        <Button leadingIcon={IconStopwatch} disabled={!!task} onClick={newRun}>
          Nueva prueba
        </Button>
      </PageHeader>

      {runs.length >= 2 ? (
        <Compare runs={runs} />
      ) : (
        <Empty title={runs.length === 0 ? "Aún no hay pruebas" : "Haz una segunda prueba para comparar"}>
          Cada prueba abre Realistic Street Soccer y mide durante 30 segundos el ping real, los FPS, los tirones y la memoria. Con dos pruebas ves el antes y el después lado a lado.
        </Empty>
      )}

      {exps.length > 0 ? (
        <div className="mt-8">
          <h2 className="mb-3 text-[15px] font-semibold text-ink">Experimentos A/B en tu PC</h2>
          <Card className="divide-y divide-line">
            {exps.map((e) => {
              const v = verdictUI[e.verdict] ?? verdictUI["no-effect"];
              return (
                <div key={e.id} className="flex items-center gap-4 px-5 py-4">
                  <span className="flex w-[104px] shrink-0">
                    <Pill tone={v.tone}>{v.label}</Pill>
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-body-medium text-ink">
                      {e.title || titles[e.tweak] || e.tweak}
                      {e.stress ? <span className="ml-2 text-caption-1-medium text-ink-3">con la CPU cargada</span> : null}
                    </p>
                    <p className="truncate text-caption-1-medium text-ink-3">{e.summary}</p>
                  </div>
                  <span className="text-caption-1-medium text-ink-3">{when(e.at)}</span>
                </div>
              );
            })}
          </Card>
        </div>
      ) : null}

      {runs.length > 0 ? (
        <div className="mt-8">
          <h2 className="mb-3 text-[15px] font-semibold text-ink">Historial</h2>
          <Card>
            <div className="grid grid-cols-[1.4fr_repeat(4,0.7fr)_auto] gap-x-4 px-5 py-3">
              {["Prueba", "Ping", "FPS", "1 % low", "Tirones/min", ""].map((h) => (
                <span key={h} className={cx("label-caps text-ink-3", h && h !== "Prueba" && "text-right")}>
                  {h}
                </span>
              ))}
            </div>
            {runs.map((r) => (
              <div key={r.id} className="grid grid-cols-[1.4fr_repeat(4,0.7fr)_auto] items-center gap-x-4 px-5 py-3 shadow-[inset_0_1px_0_var(--color-line)]">
                <div className="min-w-0">
                  <p className="truncate text-body-medium text-ink">{r.label}</p>
                  <p className="truncate text-caption-1-medium text-ink-3">
                    {when(r.at)} · {r.serverIp}
                  </p>
                </div>
                <span className="num text-right text-[19px] font-semibold text-ink">{r.window.net.confident ? num(r.window.net.rttMs) : "—"}</span>
                <span className="num text-right text-[19px] font-semibold text-ink">{num(r.window.frames.avgFps)}</span>
                <span className="num text-right text-[19px] font-semibold text-ink">{num(r.window.frames.low1Fps)}</span>
                <span className="num text-right text-[19px] font-semibold text-ink">{num(r.window.frames.stutterPerMin)}</span>
                <Button
                  iconOnly
                  variant="ghost"
                  size="small"
                  leadingIcon={IconTrash}
                  aria-label="Borrar prueba"
                  onClick={async () => {
                    await api.removeRun(r.id);
                    play("off");
                    load();
                  }}
                />
              </div>
            ))}
          </Card>
        </div>
      ) : null}
      <div className="h-4" />
      <Eyebrow className="mt-6 text-ink-4">Los resultados se guardan en C:\ProgramData\Toque</Eyebrow>
    </div>
  );
}
