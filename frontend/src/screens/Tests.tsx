import { useCallback, useEffect, useMemo, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import { api, errorText } from "@/lib/api";
import { improvement, metricMeta, num, pct, when } from "@/lib/format";
import { play } from "@/lib/sound";
import type { ExperimentGroup, ExperimentRecord, RunRecord } from "@/lib/types";
import { useApp } from "@/state";
import { Card, Empty, PageHeader, Pill } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { SegmentedControl, SegmentedControlItem } from "@/components/ui/segmented-control";
import { EffectBar, EvidenceStrip, signedPct } from "@/components/Evidence";
import type { StripPoint } from "@/components/Evidence";
import { IconChevronDown, IconStopwatch, IconTrash } from "@/components/icons";

type Row = { key: string; get: (r: RunRecord) => number | null };

const rows: Row[] = [
  { key: "rtt", get: (r) => (r.window.net.confident ? r.window.net.rttMs : null) },
  { key: "jitter", get: (r) => (r.window.net.confident ? r.window.net.jitterMs : null) },
  { key: "fps", get: (r) => r.window.frames.avgFps },
  { key: "low1", get: (r) => r.window.frames.low1Fps },
  { key: "low01", get: (r) => r.window.frames.low01Fps },
  { key: "stutterMin", get: (r) => r.window.frames.stutterPerMin },
  { key: "latMs", get: (r) => r.window.frames.renderLatencyMs },
  { key: "inputLatency", get: (r) => (r.window.frames.inputEvents >= 20 ? r.window.frames.inputLatencyMs : null) },
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
        <span className="w-24 text-right text-caption-1-medium text-ink-3">Antes</span>
        <span className="w-24 text-right text-caption-1-medium text-ink">Después</span>
        <span className="w-24 text-right text-caption-1-medium text-ink-3">Cambio</span>
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
        Una sola partida puede variar por el servidor o por los jugadores en pantalla. Para saber si un cambio mejora algo, usa "Probar en mi PC" en Optimizar, que alterna con y sin el cambio varias veces.
      </p>
    </Card>
  );
}

type Bucket = "improves" | "unconfirmed" | "no-effect" | "worse";

const verdictUI: Record<string, { label: string; tone: "volt" | "neutral" | "bad" | "warn"; bucket: Bucket }> = {
  improves: { label: "Mejora", tone: "volt", bucket: "improves" },
  tradeoff: { label: "Mejora con coste", tone: "warn", bucket: "improves" },
  unconfirmed: { label: "No se confirmó", tone: "warn", bucket: "unconfirmed" },
  "no-effect": { label: "Sin efecto", tone: "neutral", bucket: "no-effect" },
  worse: { label: "Empeora", tone: "bad", bucket: "worse" },
};

const order: Bucket[] = ["improves", "unconfirmed", "no-effect", "worse"];

interface Group {
  key: string;
  title: string;
  tests: ExperimentRecord[];
  verdict: string;
  bucket: Bucket;
  metric: string;
  points: StripPoint[];
  windows: number;
  stress: boolean;
  headline: { label: string; unit: string; digits: number; a: number; b: number; d: number } | null;
}

function buildGroups(groups: ExperimentGroup[]): Group[] {
  return groups
    .filter((g) => g.tests?.length)
    .map((g) => {
      const latest = g.tests[0];
      const metric = (latest.primary ?? []).find((k) => metricMeta[k]) ?? "latMs";
      const points: StripPoint[] = [];
      for (const t of g.tests) {
        let a = 0;
        let b = 0;
        for (const w of t.windows ?? []) {
          const index = w.arm === "B" ? ++b : ++a;
          const v = w.values?.[metric];
          if (w.valid && typeof v === "number" && Number.isFinite(v)) points.push({ arm: w.arm === "B" ? "B" : "A", value: v, at: t.at, index });
        }
      }
      const meta = metricMeta[metric];
      const c = g.pooled?.[metric] ?? latest.compare[metric];
      const ui = verdictUI[g.verdict] ?? verdictUI["no-effect"];
      return {
        key: `${g.tweak}|${g.stress}`,
        title: g.title || g.tweak,
        tests: g.tests,
        verdict: g.verdict,
        bucket: ui.bucket,
        metric,
        points,
        windows: g.tests.reduce((n, t) => n + t.valid, 0),
        stress: !!g.stress,
        headline: meta && c && c.a ? { label: meta.label, unit: meta.unit, digits: meta.digits + 1, a: c.a, b: c.b, d: c.pct } : null,
      };
    })
    .sort((x, y) => order.indexOf(x.bucket) - order.indexOf(y.bucket) || y.tests[0].at.localeCompare(x.tests[0].at));
}

function EffectTable({ test }: { test: ExperimentRecord }) {
  const primary = test.primary ?? [];
  const keys = Object.keys(test.compare)
    .filter((k) => metricMeta[k] && Number.isFinite(test.compare[k].pct) && test.compare[k].a !== 0)
    .filter((k) => primary.includes(k) || test.compare[k].conclusive)
    .sort((x, y) => (primary.includes(x) ? 0 : 1) - (primary.includes(y) ? 0 : 1));
  if (keys.length === 0) return null;
  const scale = Math.min(60, Math.max(5, ...keys.map((k) => Math.max(Math.abs(test.compare[k].low), Math.abs(test.compare[k].high)))));
  return (
    <div className="grid grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1.6fr)_72px] items-center gap-x-4 gap-y-1">
      <span className="text-caption-1-medium text-ink-3">Métrica</span>
      <span className="text-right text-caption-1-medium text-ink-3">Sin → con</span>
      <span className="text-center text-caption-1-medium text-ink-3">Cambio con su intervalo del 95 %</span>
      <span className="text-right text-caption-1-medium text-ink-3">Cambio</span>
      {keys.map((k) => {
        const c = test.compare[k];
        const meta = metricMeta[k];
        const digits = meta.digits >= 2 ? meta.digits : meta.digits + 1;
        const good = improvement(k, c.pct) > 0;
        const tone = !c.conclusive ? "neutral" : good ? "volt" : "bad";
        return (
          <div key={k} className="contents">
            <span className="truncate py-1.5 text-body-regular text-ink-2">
              {meta.label}
              {primary.includes(k) ? <span className="ml-1.5 text-caption-1-medium text-ink-4">principal</span> : null}
            </span>
            <span className="num py-1.5 text-right text-[15px] font-semibold text-ink">
              {num(c.a, digits)} → {num(c.b, digits)}
              <span className="ml-1 font-sans text-[11px] font-medium text-ink-4">{meta.unit}</span>
            </span>
            <EffectBar low={c.low} high={c.high} value={c.pct} tone={tone} scale={scale} />
            <span className="num py-1.5 text-right text-[15px] font-semibold text-ink">
              {signedPct(c.pct)}
              <span className="sr-only">{c.conclusive ? (good ? " mejora" : " empeora") : " sin efecto claro"}</span>
            </span>
          </div>
        );
      })}
    </div>
  );
}

function GroupCard({ g }: { g: Group }) {
  const [open, setOpen] = useState(false);
  const ui = verdictUI[g.verdict] ?? verdictUI["no-effect"];
  const h = g.headline;
  return (
    <Card className="p-5">
      <div className="grid grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)] gap-8">
        <div className="min-w-0">
          <div className="flex items-center gap-2.5">
            <Pill tone={ui.tone}>{ui.label}</Pill>
            <span className="text-caption-1-medium text-ink-3">
              {g.tests.length === 1 ? "1 prueba" : `${g.tests.length} pruebas`} · {g.windows} ventanas · {when(g.tests[0].at)}
            </span>
          </div>
          <h3 className="mt-3 text-[17px] font-semibold text-ink">
            {g.title}
            {g.stress ? <span className="ml-2 text-caption-1-medium font-medium text-ink-3">con la CPU cargada</span> : null}
          </h3>
          {h ? (
            <p className="mt-3 flex items-baseline gap-2">
              <span className="num text-[30px] leading-none font-semibold text-ink">{signedPct(h.d)}</span>
              <span className="text-body-regular text-ink-2">
                {h.label}, {g.tests.length > 1 ? `juntando ${g.tests.length} pruebas` : "en una prueba"}
              </span>
            </p>
          ) : null}
          <p className="mt-2 text-caption-1-medium text-ink-3">{g.tests[0].summary}</p>
        </div>
        <div className="min-w-0">
          <p className="mb-1 text-caption-1-medium text-ink-3">{metricMeta[g.metric]?.label} de cada ventana</p>
          {g.points.length > 0 ? <EvidenceStrip metric={g.metric} points={g.points} /> : <p className="mt-6 text-caption-1-medium text-ink-3">Sin ventanas válidas para dibujar.</p>}
        </div>
      </div>
      <button
        onClick={() => {
          play("tap");
          setOpen((v) => !v);
        }}
        aria-expanded={open}
        className="mt-3 flex items-center gap-1.5 rounded-md text-caption-1-semibold text-ink-2 outline-none hover:text-ink focus-visible:ring-2 focus-visible:ring-border-focus-ring"
      >
        {open ? "Ocultar detalle" : "Ver el detalle de cada prueba"}
        <IconChevronDown className={cx("size-3.5 transition-transform duration-200", open && "rotate-180")} />
      </button>
      <AnimatePresence initial={false}>
        {open ? (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.22, ease: [0.22, 1, 0.36, 1] }}
            className="overflow-hidden"
          >
            <div className="mt-4 flex flex-col gap-5">
              {g.tests.map((t) => {
                const tu = verdictUI[t.verdict] ?? verdictUI["no-effect"];
                return (
                  <div key={t.id} className="rounded-xl bg-surface-2 p-4">
                    <div className="mb-3 flex items-center gap-2.5">
                      <Pill tone={tu.tone}>{tu.label}</Pill>
                      <span className="text-caption-1-medium text-ink-3">
                        {when(t.at)} · {t.valid} ventanas válidas
                      </span>
                    </div>
                    <EffectTable test={t} />
                  </div>
                );
              })}
            </div>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </Card>
  );
}

type Filter = "all" | Bucket;

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div className="rounded-2xl bg-surface-1 px-5 py-4 shadow-[inset_0_0_0_1px_var(--color-line)]">
      <p className="num text-[34px] leading-none font-semibold text-ink">{value}</p>
      <p className="mt-2 text-caption-1-medium text-ink-3">{label}</p>
    </div>
  );
}

export function Tests() {
  const { run, task, toast } = useApp();
  const [runs, setRuns] = useState<RunRecord[]>([]);
  const [raw, setRaw] = useState<ExperimentGroup[]>([]);
  const [filter, setFilter] = useState<Filter>("all");
  const load = useCallback(async () => {
    try {
      const [r, g] = await Promise.all([api.runs(), api.experimentGroups()]);
      setRuns(r ?? []);
      setRaw(g ?? []);
    } catch (e) {
      toast({ tone: "error", title: "No pude leer el historial", detail: errorText(e) });
    }
  }, [toast]);
  useEffect(() => {
    load();
  }, [load]);
  const groups = useMemo(() => buildGroups(raw), [raw]);
  const count = (b: Bucket) => groups.filter((g) => g.bucket === b).length;
  const shown = filter === "all" ? groups : groups.filter((g) => g.bucket === filter);
  const windows = groups.reduce((n, g) => n + g.windows, 0);
  const confirmed = groups.filter((g) => g.bucket === "improves" && g.tests.length > 1).length;
  const newRun = async () => {
    const rec = await run("benchmark", "Midiendo una partida", () => api.benchmark(""));
    if (rec) {
      toast({ tone: "ok", title: "Partida medida", detail: `${num(rec.window.frames.avgFps)} FPS, ${num(rec.window.net.rttMs)} ms de ping` });
      load();
    }
  };
  return (
    <div>
      <PageHeader eyebrow="Pruebas" title="Lo que se midió en tu PC">
        <Button leadingIcon={IconStopwatch} disabled={!!task} onClick={newRun}>
          Medir una partida
        </Button>
      </PageHeader>

      {groups.length > 0 ? (
        <>
          <div className="grid grid-cols-3 gap-4">
            <Stat value={num(groups.length)} label="optimizaciones probadas dentro del juego" />
            <Stat value={num(windows)} label="ventanas de 25 segundos medidas" />
            <Stat value={num(confirmed)} label={confirmed === 1 ? "mejora confirmada con su réplica" : "mejoras confirmadas con su réplica"} />
          </div>

          <div className="mt-8 flex items-center justify-between gap-4">
            <h2 className="text-[17px] font-semibold text-ink">Experimentos A/B</h2>
            <SegmentedControl
              selectedKeys={[filter]}
              onSelectionChange={(keys) => {
                const k = [...keys][0] as Filter | undefined;
                if (k) {
                  play("tap");
                  setFilter(k);
                }
              }}
              aria-label="Filtrar experimentos"
            >
              <SegmentedControlItem id="all">Todos {groups.length}</SegmentedControlItem>
              <SegmentedControlItem id="improves">Mejoran {count("improves")}</SegmentedControlItem>
              <SegmentedControlItem id="unconfirmed">No se confirmó {count("unconfirmed")}</SegmentedControlItem>
              <SegmentedControlItem id="no-effect">Sin efecto {count("no-effect")}</SegmentedControlItem>
              <SegmentedControlItem id="worse">Empeoran {count("worse")}</SegmentedControlItem>
            </SegmentedControl>
          </div>
          <p className="mt-2 max-w-[78ch] text-body-regular text-ink-2">
            Cada punto es una ventana medida dentro de Realistic Street Soccer, sin el cambio en gris y con el cambio en verde. Cuando las dos nubes no se tocan, la diferencia no es casualidad.
          </p>
          <div className="mt-4 flex flex-col gap-4">
            {shown.map((g) => (
              <GroupCard key={g.key} g={g} />
            ))}
            {shown.length === 0 ? <p className="py-8 text-center text-body-regular text-ink-3">Ningún experimento con este resultado.</p> : null}
          </div>
        </>
      ) : null}

      <h2 className="mt-10 mb-3 text-[17px] font-semibold text-ink">Tus partidas medidas</h2>
      {runs.length >= 2 ? (
        <Compare runs={runs} />
      ) : (
        <Empty title={runs.length === 0 ? "Aún no has medido ninguna partida" : "Mide otra partida para comparar"}>
          Cada medición abre Realistic Street Soccer y registra durante 30 segundos el ping real, los FPS, los tirones y la memoria. Si juegas mientras mide, también registra la latencia de entrada, desde que pulsas hasta que se ve en pantalla. Con dos mediciones ves el antes y el después lado a lado.
        </Empty>
      )}

      {runs.length > 0 ? (
        <Card className="mt-4">
          <div className="grid grid-cols-[1.4fr_repeat(4,0.7fr)_auto] gap-x-4 px-5 py-3">
            {["Partida", "Ping", "FPS", "1 % low", "Tirones/min", ""].map((h) => (
              <span key={h} className={cx("text-caption-1-medium text-ink-3", h && h !== "Partida" && "text-right")}>
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
                aria-label="Borrar medición"
                onClick={async () => {
                  await api.removeRun(r.id);
                  play("off");
                  load();
                }}
              />
            </div>
          ))}
        </Card>
      ) : null}
      <p className="mt-8 text-caption-1-medium text-ink-4">Los resultados se guardan en C:\ProgramData\Toque</p>
    </div>
  );
}
