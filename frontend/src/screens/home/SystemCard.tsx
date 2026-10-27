import type { ComponentType } from "react";
import { num } from "@/lib/format";
import { play } from "@/lib/sound";
import { useApp } from "@/state";
import { Button } from "@/components/ui/button";
import { IconClock, IconCpu, IconDiagnose, IconGpu, IconPc, IconRam, IconShield, IconThreads, IconWindows } from "@/components/icons";

function Row({ icon: Icon, label, value }: { icon: ComponentType<{ className?: string }>; label: string; value: string }) {
  return (
    <div className="flex items-center gap-3 rounded-lg px-2 py-[7px] odd:bg-background-inner-default">
      <Icon className="size-4 shrink-0 text-ink-3" />
      <span className="w-[68px] shrink-0 text-caption-1-medium text-ink-3">{label}</span>
      <span className="min-w-0 truncate text-body-2-medium text-ink">{value}</span>
    </div>
  );
}

function uptime(min: number) {
  if (min < 60) return `${num(min)} min`;
  const h = Math.floor(min / 60);
  return h < 48 ? `${h} h ${min % 60} min` : `${Math.floor(h / 24)} días`;
}

export function SystemCard() {
  const { overview, go } = useApp();
  const s = overview?.system;
  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
        <IconPc className="size-4 text-ink-3" />
        Tu PC
      </p>
      <div className="mt-3 flex flex-col">
        <Row icon={IconCpu} label="CPU" value={s ? s.cpu.replace(/\s+CPU.*$/i, "").replace(/^Intel\(R\) Core\(TM\)\s*/i, "").replace(/^AMD\s+/i, "") : "—"} />
        <Row icon={IconThreads} label="Hilos" value={s ? `${num(s.threads)} hilos` : "—"} />
        <Row icon={IconGpu} label="Gráfica" value={s?.gpu.replace(/^NVIDIA GeForce\s*/i, "") ?? "—"} />
        <Row icon={IconRam} label="Memoria" value={s ? `${num(s.ramGb, 0)} GB` : "—"} />
        <Row icon={IconPc} label="Monitor" value={s ? `${num(s.refreshHz)} Hz` : "—"} />
        <Row icon={IconWindows} label="Sistema" value={s?.os.replace(/^Microsoft\s+/i, "") ?? "—"} />
        <Row icon={IconClock} label="Encendido" value={overview ? uptime(overview.uptimeMin) : "—"} />
        <Row icon={IconShield} label="Permisos" value={overview ? (overview.elevated ? "Administrador" : "Limitados") : "—"} />
      </div>
      <div className="mt-auto pt-4">
        <Button
          variant="secondary"
          size="small"
          className="w-full"
          leadingIcon={IconDiagnose}
          onClick={() => {
            play("tap");
            go("diagnose");
          }}
        >
          Diagnosticar qué te frena
        </Button>
      </div>
    </section>
  );
}
