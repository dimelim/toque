import { useCallback, useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { api, errorText } from "@/lib/api";
import { play } from "@/lib/sound";
import type { WindowsOption } from "@/lib/types";
import { useApp } from "@/state";
import { Card, Eyebrow, PageHeader } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Switch } from "@/components/ui/switch";
import { IconRevert, IconWindows } from "@/components/icons";

const groups = [
  { id: "datos", title: "Datos que Windows envía" },
  { id: "anuncios", title: "Publicidad y sugerencias" },
  { id: "funciones", title: "Funciones que puedes quitar" },
];

function Row({ o, disabled, onChange }: { o: WindowsOption; disabled: boolean; onChange: (on: boolean) => void }) {
  const already = o.applied && !o.managed;
  const [open, setOpen] = useState(false);
  return (
    <div className="flex items-start justify-between gap-6 py-4 shadow-[inset_0_-1px_0_var(--color-line)] last:shadow-none">
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-body-medium text-ink">{o.title}</p>
          {already ? (
            <Chip variant="caption" color="neutral">
              Ya lo tenías así
            </Chip>
          ) : null}
          {!o.applicable ? (
            <Chip variant="caption" color="yellow">
              {o.note || "Necesita administrador"}
            </Chip>
          ) : null}
        </div>
        <p className="mt-1 max-w-[70ch] text-body-regular text-ink-2">{o.what}</p>
        <button className="mt-2 text-caption-1-medium text-ink-3 underline-offset-2 outline-none hover:text-ink-2 hover:underline focus-visible:underline" onClick={() => setOpen(!open)} aria-expanded={open}>
          {open ? "Ocultar qué cambia" : "Ver qué cambia"}
        </button>
        <AnimatePresence initial={false}>
          {open ? (
            <motion.ul
              className="mt-2 flex flex-col gap-1 overflow-hidden rounded-lg bg-surface-1 px-3 py-2 font-mono text-[11.5px] leading-relaxed text-ink-3"
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
            >
              {o.changes.map((c) => (
                <li key={c} className="break-all">
                  {c}
                </li>
              ))}
            </motion.ul>
          ) : null}
        </AnimatePresence>
      </div>
      <Switch size="md" className="mt-0.5 shrink-0" isSelected={o.applied} isDisabled={disabled || !o.applicable || already} onChange={onChange} aria-label={o.title} />
    </div>
  );
}

export function WindowsTab() {
  const { task, toast, refresh } = useApp();
  const [items, setItems] = useState<WindowsOption[] | null>(null);
  const [busy, setBusy] = useState(false);
  const load = useCallback(async () => {
    try {
      setItems(await api.windowsOptions());
    } catch (e) {
      toast({ tone: "error", title: "No pude leer los ajustes de Windows", detail: errorText(e) });
    }
  }, [toast]);
  useEffect(() => {
    load();
  }, [load]);

  const change = async (fn: () => Promise<WindowsOption[]>, on: boolean, done?: string) => {
    setBusy(true);
    try {
      setItems(await fn());
      play(on ? "on" : "off");
      if (done) toast({ tone: "ok", title: done, detail: on ? "Algunos cambios, como la búsqueda o las recomendaciones de Inicio, se notan al volver a iniciar sesión." : undefined });
    } catch (e) {
      play("error");
      toast({ tone: "error", title: "No pude cambiar Windows", detail: errorText(e) });
      load();
    } finally {
      setBusy(false);
      refresh();
    }
  };

  const disabled = busy || !!task || !items;
  const pending = (items ?? []).filter((o) => o.applicable && !o.applied).length;
  const mine = (items ?? []).filter((o) => o.managed).length;
  return (
    <div>
      <PageHeader eyebrow="Optimizar" title="Windows más limpio">
        <Button variant="secondary" leadingIcon={IconRevert} disabled={disabled || mine === 0} onClick={() => change(() => api.setAllWindowsOptions(false), false, "Windows vuelve a como estaba")}>
          Deshacer todo
        </Button>
        <Button leadingIcon={IconWindows} disabled={disabled || pending === 0} onClick={() => change(() => api.setAllWindowsOptions(true), true, "Windows envía menos datos y te muestra menos sugerencias")}>
          {pending === 0 ? "Todo aplicado" : `Aplicar ${pending}`}
        </Button>
      </PageHeader>
      <Card className="mb-8 p-5">
        <Eyebrow>Qué hace</Eyebrow>
        <p className="mt-2 max-w-[78ch] text-body-regular text-ink-2">
          Quita la publicidad, las sugerencias, funciones que no usas y el envío de datos opcionales de Windows. No sube los FPS por sí solo. Cada cambio se guarda y se deshace con un clic, desde Ajustes o al desinstalar Toque.
        </p>
      </Card>
      <div className="flex flex-col gap-4">
        {groups.map((g) => (
          <Card key={g.id} className="px-5 py-2">
            <h2 className="pt-3 text-[15px] font-semibold text-ink">{g.title}</h2>
            {(items ?? [])
              .filter((o) => o.group === g.id)
              .map((o) => (
                <Row key={o.id} o={o} disabled={disabled} onChange={(on) => change(() => api.setWindowsOption(o.id, on), on)} />
              ))}
          </Card>
        ))}
      </div>
    </div>
  );
}
