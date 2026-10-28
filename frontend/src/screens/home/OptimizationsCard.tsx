import { api } from "@/lib/api";
import { metricMeta, pct } from "@/lib/format";
import { t } from "@/lib/i18n";
import { play } from "@/lib/sound";
import type { Optimization } from "@/lib/types";
import { useApp } from "@/state";
import { Chip } from "@/components/ui/chip";
import { SwitchCard } from "@/components/ui/switch-card";
import { IconArrowRight, IconOptimize } from "@/components/icons";

function effect(o: Optimization) {
  const mine = o.mine;
  if (!mine) return null;
  const key = (mine.primary ?? []).find((k) => mine.compare[k]?.conclusive) ?? (mine.primary ?? [])[0];
  const c = key ? mine.compare[key] : null;
  const meta = metricMeta(key);
  if (!c || !meta) return null;
  return `${meta.label} ${pct(c.pct, Math.abs(c.pct) < 10 ? 1 : 0)}`;
}

function VerdictBar({ all }: { all: Optimization[] }) {
  const parts = [
    { key: t("mejoran"), n: all.filter((o) => o.verdict === "proven" || o.verdict === "tradeoff").length, bar: "bg-volt", dot: "bg-volt" },
    { key: t("según tu caso"), n: all.filter((o) => o.verdict === "situational").length, bar: "bg-status-cyan-text", dot: "bg-status-cyan-text" },
    { key: t("descartadas"), n: all.filter((o) => o.verdict === "rejected" || o.verdict === "pending").length, bar: "bg-surface-4", dot: "bg-ink-4" },
  ].filter((p) => p.n > 0);
  return (
    <div className="mt-4 rounded-xl bg-background-inner-default p-3 shadow-card">
      <p className="text-caption-1-medium text-ink-3">{t("Lo que dieron las pruebas en Realistic Street Soccer")}</p>
      <div className="mt-2.5 flex h-2 gap-0.5 overflow-hidden rounded-full">
        {parts.map((p) => (
          <span key={p.key} className={`h-full rounded-full ${p.bar}`} style={{ flexGrow: p.n }} />
        ))}
      </div>
      <div className="mt-2.5 flex flex-wrap gap-x-3 gap-y-1">
        {parts.map((p) => (
          <span key={p.key} className="flex items-center gap-1.5 text-caption-1-medium text-ink-2">
            <span className={`size-2 rounded-full ${p.dot}`} />
            <span className="num text-[14px] font-semibold text-ink">{p.n}</span>
            {p.key}
          </span>
        ))}
      </div>
    </div>
  );
}

export function OptimizationsCard({ items, all, onChange }: { items: Optimization[]; all: Optimization[]; onChange: () => void }) {
  const total = all.length;
  const { go, run, task, toast } = useApp();
  const toggle = async (o: Optimization, on: boolean) => {
    play(on ? "on" : "off");
    const res = await run(on ? "apply" : "revert", on ? t("Aplicando {title}", { title: o.title }) : t("Quitando {title}", { title: o.title }), () => (on ? api.apply(o.id) : api.revert(o.id)));
    if (res) {
      toast({ tone: "ok", title: on ? t("{title} aplicada", { title: o.title }) : t("{title} quitada", { title: o.title }), detail: o.kind === "relaunch" ? t("Se nota la próxima vez que abras Roblox.") : undefined });
      onChange();
    }
  };
  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
          <IconOptimize className="size-4 text-ink-3" />
          {t("Optimizaciones probadas")}
        </p>
        <Chip variant="caption" color="lime">
          {t("{n} de {total}", { n: items.length, total })}
        </Chip>
      </div>
      <div className="mt-3 flex flex-col gap-2">
        {items.map((o) => (
          <SwitchCard
            key={o.id}
            title={o.title}
            description={[effect(o), o.verdict === "tradeoff" ? t("con coste visual") : null].filter(Boolean).join(" · ")}
            size="sm"
            isSelected={o.applied}
            isDisabled={!!task || !o.applicable || (o.applied && !o.managed)}
            onChange={(on) => toggle(o, on)}
          />
        ))}
      </div>
      <VerdictBar all={all} />
      <button
        onClick={() => {
          play("tap");
          go("optimize");
        }}
        className="mt-auto flex items-center gap-1.5 self-start rounded-md pt-4 text-caption-1-semibold text-ink-2 outline-none hover:text-ink focus-visible:ring-2 focus-visible:ring-border-focus-ring"
      >
        {t("Ver las {total} con sus pruebas", { total })}
        <IconArrowRight className="size-3.5" />
      </button>
    </section>
  );
}
