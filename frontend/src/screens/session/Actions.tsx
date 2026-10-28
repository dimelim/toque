import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import { improvement, metricMeta, num, when } from "@/lib/format";
import type { ExperimentRecord, SessionAction, SessionProc } from "@/lib/types";
import { Card, DeltaChip, Pill } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Switch } from "@/components/ui/switch";
import { IconFlask } from "@/components/icons";

const statusLabel: Record<SessionAction["status"], string> = {
  off: "Apagada",
  waiting: "En espera",
  active: "Activa",
};

function Indicator({ status }: { status: SessionAction["status"] }) {
  return (
    <span className="flex items-center gap-2 text-caption-1-medium text-ink-3">
      <span className={cx("relative size-2 rounded-full", status === "active" ? "live-dot bg-volt text-volt" : status === "waiting" ? "bg-volt-line" : "bg-ink-4")} />
      <AnimatePresence mode="wait" initial={false}>
        <motion.span key={status} initial={{ opacity: 0, y: 3 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} transition={{ duration: 0.15 }} className={status === "active" ? "text-volt" : undefined}>
          {statusLabel[status]}
        </motion.span>
      </AnimatePresence>
    </span>
  );
}

function ProcChips({ procs }: { procs: SessionProc[] }) {
  if (procs.length === 0) return null;
  return (
    <div className="mt-3 flex flex-wrap gap-1.5">
      {procs.map((p) => (
        <Chip key={p.name} variant="caption" color="neutral">
          {p.name}
          <span className="num ml-1.5 text-ink-3">{num(p.cpu, p.cpu >= 10 ? 0 : 1)} %</span>
        </Chip>
      ))}
    </div>
  );
}

function Result({ exp }: { exp: ExperimentRecord }) {
  const keys = (exp.primary ?? []).filter((k) => metricMeta[k] && exp.compare[k]);
  const tone = exp.verdict === "improves" ? "volt" : exp.verdict === "tradeoff" ? "warn" : exp.verdict === "worse" ? "bad" : "neutral";
  const label = exp.verdict === "improves" ? "Mejora en tu PC" : exp.verdict === "tradeoff" ? "Mejora con un coste en tu PC" : exp.verdict === "worse" ? "Empeora en tu PC" : "Sin efecto claro en tu PC";
  return (
    <motion.div initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }} transition={{ type: "spring", duration: 0.4, bounce: 0 }} className="mt-4 rounded-xl bg-surface-2 p-4">
      <div className="flex flex-wrap items-center gap-1.5">
        <Pill tone={tone}>{label}</Pill>
        {keys.slice(0, 3).map((k) => (
          <DeltaChip key={k} metric={k} deltaPct={exp.compare[k].pct} conclusive={exp.compare[k].conclusive && improvement(k, exp.compare[k].pct) !== 0} />
        ))}
        <span className="text-caption-1-medium text-ink-3">{when(exp.at)}</span>
      </div>
      <p className="mt-2 text-body-regular text-ink-2">{exp.summary}</p>
    </motion.div>
  );
}

export function ActionCard({
  action,
  busy,
  taskActive,
  onToggle,
  onVerify,
}: {
  action: SessionAction;
  busy: boolean;
  taskActive: boolean;
  onToggle: (on: boolean) => void;
  onVerify: () => void;
}) {
  const verifiable = action.verifyId !== "";
  const experiment = action.experiment;
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2.5">
            <h3 className="text-[15px] font-semibold text-ink">{action.title}</h3>
            {!action.tested ? (
              <Chip variant="caption" color="yellow">
                Sin probar
              </Chip>
            ) : null}
            <Indicator status={action.status} />
          </div>
          <p className="mt-1.5 text-body-regular text-ink-2">{action.what}</p>
          {action.detail ? <p className="mt-2 text-caption-1-medium text-ink-3">{action.detail}</p> : null}
          <ProcChips procs={action.procs} />
          {verifiable ? (
            <div className="mt-4">
              <Button size="small" variant="ghost" leadingIcon={IconFlask} disabled={taskActive} onClick={onVerify}>
                Probar en mi PC
              </Button>
            </div>
          ) : null}
          {verifiable && experiment ? <Result exp={experiment} /> : null}
        </div>
        <Switch size="md" isSelected={action.enabled} isDisabled={busy || taskActive} onChange={onToggle} aria-label={action.title} />
      </div>
    </Card>
  );
}
