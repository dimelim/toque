import { useCallback, useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import { api, errorText } from "@/lib/api";
import { improvement, metricMeta, num, when } from "@/lib/format";
import { play } from "@/lib/sound";
import type { Optimization } from "@/lib/types";
import { useApp } from "@/state";
import { Card, DeltaChip, Eyebrow, PageHeader, Pill } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { IconChevronDown, IconFlask, IconOptimize, IconRevert } from "@/components/icons";

const kindLabel: Record<string, string> = {
  live: "Al instante",
  relaunch: "Al reabrir Roblox",
  reboot: "Tras reiniciar el PC",
};

const categoryLabel: Record<string, string> = {
  memoria: "Memoria",
  sistema: "Sistema",
  fps: "Gráficos",
  red: "Red",
};

function evidenceKeys(o: Optimization) {
  const mine = o.mine;
  if (!mine) return [];
  const primary = mine.primary ?? [];
  return Object.keys(mine.compare)
    .filter((k) => metricMeta[k] && mine.compare[k].conclusive && Math.abs(mine.compare[k].pct) >= 1)
    .sort((a, b) => {
      const pa = primary.includes(a) ? 0 : 1;
      const pb = primary.includes(b) ? 0 : 1;
      if (pa !== pb) return pa - pb;
      return Math.abs(improvement(b, mine.compare[b].pct)) - Math.abs(improvement(a, mine.compare[a].pct));
    });
}

function Evidence({ o }: { o: Optimization }) {
  const mine = o.mine;
  const keys = evidenceKeys(o);
  if (mine && keys.length > 0) {
    return (
      <div className="flex flex-wrap items-center gap-1.5">
        <span className="text-caption-1-medium text-ink-3">En tu PC</span>
        {keys.slice(0, 3).map((k) => (
          <DeltaChip key={k} metric={k} deltaPct={mine.compare[k].pct} />
        ))}
      </div>
    );
  }
  if (o.lab) {
    return <p className="text-caption-1-medium text-ink-3">{o.lab}</p>;
  }
  return null;
}

function Detail({ label, children }: { label: string; children: string }) {
  return (
    <div>
      <p className="text-caption-1-semibold text-ink-3">{label}</p>
      <p className="mt-1 text-body-regular text-ink-2">{children}</p>
    </div>
  );
}

function OptimizationCard({ o, onChange }: { o: Optimization; onChange: (o: Optimization) => void }) {
  const { run, task, toast } = useApp();
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const rejected = o.verdict === "rejected";
  const toggle = async (on: boolean) => {
    setBusy(true);
    try {
      const next = on ? await api.apply(o.id) : await api.revert(o.id);
      play(on ? "on" : "off");
      onChange(next);
      if (on && o.kind === "reboot") toast({ tone: "info", title: "Se aplicará al reiniciar el PC" });
      if (on && o.kind === "relaunch") toast({ tone: "info", title: "Se aplicará la próxima vez que abras Roblox" });
    } catch (e) {
      play("error");
      toast({ tone: "error", title: o.title, detail: errorText(e) });
    } finally {
      setBusy(false);
    }
  };
  const verify = async () => {
    const rec = await run("verify", `Probando "${o.title}" · no toques el PC`, () => api.verify(o.id));
    if (rec) {
      toast({ tone: rec.verdict === "improves" ? "ok" : "info", title: rec.verdict === "improves" ? "Mejora confirmada en tu PC" : "Sin mejora clara en tu PC", detail: rec.summary });
      onChange({ ...o, mine: rec });
    }
  };
  return (
    <Card className={cx("p-5", rejected && "opacity-80")}>
      <div className="flex items-start gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-1.5">
            <Pill>{categoryLabel[o.category] ?? o.category}</Pill>
            {o.kind ? <Pill>{kindLabel[o.kind] ?? o.kind}</Pill> : null}
            {o.applied && !rejected ? <Pill tone="volt">Activa</Pill> : null}
            {!o.applicable && o.detail ? <Pill tone="warn">No aplica aquí</Pill> : null}
          </div>
          <h3 className="mt-3 text-[17px] leading-snug font-semibold text-ink">{o.title}</h3>
          <p className="mt-1 max-w-[70ch] text-body-regular text-ink-2">{o.what}</p>
          <div className="mt-3">
            <Evidence o={o} />
          </div>
        </div>
        {!rejected ? (
          <Switch
            size="md"
            isSelected={o.applied}
            isDisabled={busy || !!task || (!o.applicable && !o.applied)}
            onChange={toggle}
            aria-label={`Activar ${o.title}`}
          />
        ) : null}
      </div>
      <div className="mt-4 flex items-center gap-2">
        <Button variant="ghost" size="small" trailingIcon={IconChevronDown} onClick={() => setOpen((v) => !v)} aria-expanded={open}>
          {open ? "Ocultar detalles" : "Ver detalles"}
        </Button>
        <Button variant="ghost" size="small" leadingIcon={IconFlask} disabled={!!task} onClick={verify}>
          {rejected ? "Probarla igualmente en mi PC" : "Probar en mi PC"}
        </Button>
        {o.mine ? <span className="ml-auto text-caption-1-medium text-ink-3">Última prueba {when(o.mine.at)}</span> : null}
      </div>
      <AnimatePresence initial={false}>
        {open ? (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.22, ease: [0.22, 1, 0.36, 1] }}
            className="overflow-hidden"
          >
            <div className="mt-4 grid grid-cols-2 gap-x-8 gap-y-4 rounded-xl bg-surface-2 p-4">
              <Detail label="Por qué debería ayudar">{o.why}</Detail>
              <Detail label="Qué métrica afecta">{o.metric}</Detail>
              <Detail label="Puede empeorar">{o.risk}</Detail>
              <Detail label="Cómo se revierte">{o.revert}</Detail>
              {o.mine ? <Detail label="Resultado en tu PC">{`${o.mine.summary}. ${o.mine.valid} ventanas válidas de medición.`}</Detail> : null}
              {o.detail ? <Detail label="Estado en este equipo">{o.detail}</Detail> : null}
            </div>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </Card>
  );
}

function Section({ title, hint, items, onChange }: { title: string; hint: string; items: Optimization[]; onChange: (o: Optimization) => void }) {
  if (items.length === 0) return null;
  return (
    <div className="mt-8 first:mt-0">
      <div className="mb-3 flex items-baseline justify-between gap-4">
        <h2 className="text-[15px] font-semibold text-ink">{title}</h2>
        <p className="text-caption-1-medium text-ink-3">{hint}</p>
      </div>
      <div className="flex flex-col gap-3">
        {items.map((o) => (
          <OptimizationCard key={o.id} o={o} onChange={onChange} />
        ))}
      </div>
    </div>
  );
}

export function Optimize() {
  const { run, task, toast, refresh } = useApp();
  const [items, setItems] = useState<Optimization[]>([]);
  const load = useCallback(async () => {
    try {
      setItems(await api.optimizations());
    } catch (e) {
      toast({ tone: "error", title: "No pude leer las optimizaciones", detail: errorText(e) });
    }
  }, [toast]);
  useEffect(() => {
    load();
  }, [load]);
  const update = (o: Optimization) => setItems((all) => all.map((x) => (x.id === o.id ? o : x)));
  const proven = items.filter((o) => o.verdict === "proven");
  const situational = items.filter((o) => o.verdict === "situational");
  const pending = items.filter((o) => o.verdict === "pending");
  const rejected = items.filter((o) => o.verdict === "rejected");
  const ready = proven.filter((o) => o.applicable && !o.applied).length;
  return (
    <div>
      <PageHeader eyebrow="Optimizar" title="Cambios con pruebas detrás">
        <Button
          variant="secondary"
          leadingIcon={IconRevert}
          disabled={!!task}
          onClick={async () => {
            await run("revert", "Revirtiendo todos los cambios", () => api.revertAll());
            toast({ tone: "ok", title: "Todo vuelve a como estaba" });
            load();
            refresh();
          }}
        >
          Revertir todo
        </Button>
        <Button
          leadingIcon={IconOptimize}
          disabled={!!task || ready === 0}
          onClick={async () => {
            const applied = await run("apply", "Aplicando las optimizaciones probadas", () => api.applyProven());
            if (applied) toast({ tone: "ok", title: applied.length ? `${applied.length} optimizaciones aplicadas` : "Ya estaban aplicadas" });
            load();
            refresh();
          }}
        >
          {proven.length === 0 ? "Nada probado todavía" : ready > 0 ? `Aplicar ${ready} probadas` : "Todo aplicado"}
        </Button>
      </PageHeader>
      <Card className="mb-8 flex items-center gap-6 p-5">
        <div className="flex-1">
          <Eyebrow>Cómo decide Toque</Eyebrow>
          <p className="mt-2 max-w-[78ch] text-body-regular text-ink-2">
            Cada cambio se mide alternando con y sin él en orden ABBA dentro de Realistic Street Soccer, varias veces seguidas. Solo pasa a "Probada" si la mejora es estadísticamente clara y no empeora otra métrica. Lo que no mejora se queda en "Descartadas" con su medición, para que veas por qué.
          </p>
        </div>
        <div className="grid grid-cols-3 gap-6 text-right">
          <div>
            <p className="num text-[34px] leading-none font-semibold text-volt">{num(proven.length)}</p>
            <p className="mt-1 text-caption-1-medium text-ink-3">probadas</p>
          </div>
          <div>
            <p className="num text-[34px] leading-none font-semibold text-ink-2">{num(pending.length + situational.length)}</p>
            <p className="mt-1 text-caption-1-medium text-ink-3">en estudio</p>
          </div>
          <div>
            <p className="num text-[34px] leading-none font-semibold text-ink-3">{num(rejected.length)}</p>
            <p className="mt-1 text-caption-1-medium text-ink-3">descartadas</p>
          </div>
        </div>
      </Card>
      <Section title="Probadas" hint="Mejoran de forma medible en pruebas reales" items={proven} onChange={update} />
      <Section title="Según tu caso" hint="Ayudan solo en ciertas condiciones" items={situational} onChange={update} />
      <Section title="En estudio" hint="Aún sin una mejora demostrada" items={pending} onChange={update} />
      <Section title="Descartadas" hint="Las probamos y no cambian nada" items={rejected} onChange={update} />
    </div>
  );
}
