import { useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { play } from "@/lib/sound";
import { remember, useApp } from "@/state";
import type { Screen } from "@/state";
import { Card } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { IconClose } from "@/components/icons";

const key = "toque.firstSteps";

const steps: { title: string; text: string; action: string; to: Screen }[] = [
  { title: "Deja el Modo partida encendido", text: "Mientras juegas, Toque se aparta y te cuenta qué programas te frenan.", action: "Ir a Partida", to: "session" },
  { title: "Aplica lo que funciona", text: "Ajustes de Roblox con pruebas detrás y una limpieza de Windows, todo reversible.", action: "Ir a Optimizar", to: "optimize" },
  { title: "Mide tu partida", text: "Toque te enseña tus FPS reales, los tirones y el ping dentro del juego.", action: "Ir a Diagnóstico", to: "diagnose" },
];

function seen() {
  try {
    return localStorage.getItem(key) === "done";
  } catch {
    return true;
  }
}

export function FirstSteps() {
  const { go } = useApp();
  const [open, setOpen] = useState(() => !seen());
  const close = () => {
    play("tap");
    remember(key, "done");
    setOpen(false);
  };
  return (
    <AnimatePresence initial={false}>
      {open ? (
        <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, height: 0, marginBottom: -16 }} transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}>
          <Card className="p-5">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-body-medium text-ink">Primeros pasos</p>
                <p className="mt-0.5 text-caption-1-medium text-ink-3">Tres cosas para sacarle partido a Toque desde hoy.</p>
              </div>
              <button aria-label="Cerrar primeros pasos" className="grid size-8 place-items-center rounded-lg text-ink-3 outline-none transition-colors duration-150 hover:bg-surface-3 hover:text-ink focus-visible:ring-2 focus-visible:ring-border-focus-ring" onClick={close}>
                <IconClose className="size-4" />
              </button>
            </div>
            <div className="mt-4 grid grid-cols-3 gap-3">
              {steps.map((s, i) => (
                <motion.div
                  key={s.title}
                  className="flex flex-col rounded-xl bg-surface-1 p-4 shadow-[inset_0_0_0_1px_var(--color-line)]"
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.35, delay: 0.08 * i, ease: [0.22, 1, 0.36, 1] }}
                >
                  <p className="num text-[15px] font-semibold text-volt">0{i + 1}</p>
                  <p className="mt-1.5 text-body-medium text-ink">{s.title}</p>
                  <p className="mt-1 flex-1 text-body-regular text-ink-2">{s.text}</p>
                  <div className="mt-3">
                    <Button
                      size="small"
                      variant="secondary"
                      onClick={() => {
                        play("tap");
                        go(s.to);
                      }}
                    >
                      {s.action}
                    </Button>
                  </div>
                </motion.div>
              ))}
            </div>
          </Card>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}
