import { useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import { api, errorText } from "@/lib/api";
import { num } from "@/lib/format";
import { play } from "@/lib/sound";
import { useApp } from "@/state";
import type { SessionInfo, SessionNote, SessionProc } from "@/lib/types";
import { Card } from "@/components/primitives";
import { IconInfo, IconOk, IconWarn } from "@/components/icons";

function amount(note: SessionNote, p: SessionProc) {
  return note.id === "ram" ? p.ramMb : note.id === "gpu" ? p.gpu : p.cpu;
}

function amountText(note: SessionNote, value: number) {
  if (note.id !== "ram") return `${num(value, value >= 10 ? 0 : 1)} %`;
  return value >= 1024 ? `${num(value / 1024, 1)} GB` : `${num(value)} MB`;
}

function CloseButton({ name }: { name: string }) {
  const { toast } = useApp();
  const [armed, setArmed] = useState(false);
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    if (!armed) return;
    const t = setTimeout(() => setArmed(false), 3000);
    return () => clearTimeout(t);
  }, [armed]);
  const press = async () => {
    if (!armed) {
      play("tap");
      setArmed(true);
      return;
    }
    setArmed(false);
    setBusy(true);
    try {
      const n = await api.closeApp(name);
      play("off");
      toast({ tone: "ok", title: `Cerrando ${name}`, detail: n === 1 ? "Le pedí que cerrara su ventana, como si pulsaras la X." : `Le pedí que cerrara sus ${n} ventanas, como si pulsaras la X.` });
    } catch (e) {
      play("error");
      toast({ tone: "error", title: `No pude cerrar ${name}`, detail: errorText(e) });
    } finally {
      setBusy(false);
    }
  };
  return (
    <button
      className={cx(
        "rounded-md px-2 py-0.5 text-caption-1-semibold outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-border-focus-ring disabled:opacity-50",
        armed ? "bg-bad-soft text-bad" : "text-ink-3 hover:bg-surface-3 hover:text-ink",
      )}
      disabled={busy}
      onClick={press}
      aria-label={armed ? `Confirmar cerrar ${name}` : `Cerrar ${name}`}
    >
      {armed ? "¿Cerrar?" : "Cerrar"}
    </button>
  );
}

function Bar({ name, value, max, text, tone, closable }: { name: string; value: number; max: number; text: string; tone: SessionNote["tone"]; closable: boolean }) {
  const ratio = Math.max(0.02, Math.min(1, value / max));
  return (
    <div className="grid grid-cols-[minmax(0,140px)_minmax(0,1fr)_56px_72px] items-center gap-3">
      <span className="truncate text-caption-1-medium text-ink-2">{name}</span>
      <span className="h-1.5 overflow-hidden rounded-full bg-surface-3">
        <span
          className={cx("block h-full w-full origin-left rounded-full transition-transform duration-700 ease-[var(--ease-out-quint)]", tone === "warn" ? "bg-warn" : "bg-volt")}
          style={{ transform: `scaleX(${ratio})` }}
        />
      </span>
      <span className="num text-right text-caption-1-medium text-ink-3">{text}</span>
      <span className="flex justify-end">{closable ? <CloseButton name={name} /> : null}</span>
    </div>
  );
}

function Note({ note, closable }: { note: SessionNote; closable: boolean }) {
  const Icon = note.tone === "warn" ? IconWarn : IconInfo;
  const max = Math.max(note.id === "ram" ? 512 : 5, ...note.procs.map((p) => amount(note, p)));
  return (
    <motion.div layout initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -4 }} transition={{ type: "spring", duration: 0.4, bounce: 0 }}>
      <Card className="p-5">
        <div className="flex items-start gap-3">
          <Icon className={cx("mt-0.5 size-[18px] shrink-0", note.tone === "warn" ? "text-warn" : "text-ink-2")} />
          <div className="min-w-0 flex-1">
            <h3 className="text-[15px] font-semibold text-ink">{note.title}</h3>
            <p className="mt-1 text-body-regular text-ink-2">{note.detail}</p>
            {note.procs.length > 0 ? (
              <div className="mt-4 flex flex-col gap-2.5">
                {note.procs.map((p) => (
                  <Bar key={p.name} name={p.name} value={amount(note, p)} max={max} text={amountText(note, amount(note, p))} tone={note.tone} closable={closable && note.id !== "defender" && note.id !== "exitlag"} />
                ))}
              </div>
            ) : null}
          </div>
        </div>
      </Card>
    </motion.div>
  );
}

export function Notes({ info }: { info: SessionInfo }) {
  const active = info.state === "active";
  return (
    <div className="flex flex-col gap-3">
      <AnimatePresence initial={false} mode="popLayout">
        {info.notes.map((n) => (
          <Note key={n.id} note={n} closable={info.state !== "paused"} />
        ))}
      </AnimatePresence>
      {info.notes.length === 0 ? (
        <div className="flex items-center gap-4 rounded-2xl px-6 py-8 shadow-[inset_0_0_0_1px_var(--color-line)]">
          <IconOk className={cx("size-6 shrink-0", active ? "text-volt" : "text-ink-4")} />
          <div>
            <p className="text-body-medium text-ink">{active ? "Nada te está frenando ahora mismo" : "Todavía no hay nada que mostrar"}</p>
            <p className="mt-1 text-body-regular text-ink-2">
              {active ? "Toque sigue mirando y te avisa aquí si algo empieza a gastar de más." : "Los datos aparecen durante tu partida. Cuando juegues, Toque te dice aquí qué consume CPU, gráfica o memoria."}
            </p>
          </div>
        </div>
      ) : null}
    </div>
  );
}
