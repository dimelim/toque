import { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { api } from "@/lib/api";
import { t } from "@/lib/i18n";
import { play } from "@/lib/sound";
import { useApp } from "@/state";
import { AppIcon } from "./Logo";
import { Button } from "./ui/button";

function sentences(notes: string) {
  return notes
    .split(/(?<=\.)\s+/)
    .map((s) => s.trim())
    .filter(Boolean);
}

export function WhatsNew() {
  const { update, toast } = useApp();
  const [version, setVersion] = useState<string | null>(null);
  const announced = useRef("");

  useEffect(() => {
    api
      .whatsNew()
      .then((w) => w.show && setVersion(w.version))
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    if (update?.state !== "ready" || announced.current === update.latest) return;
    announced.current = update.latest;
    play("success");
    toast({ tone: "info", title: t("Toque {version} está lista", { version: update.latest }), detail: t("Se instala sola al volver a abrir Toque, o ahora desde la barra lateral.") });
  }, [update, toast]);

  const open = !!version && update?.latest === version && !!update.notes;
  const close = () => {
    play("tap");
    api.seenWhatsNew().catch(() => undefined);
    setVersion(null);
  };

  return (
    <AnimatePresence>
      {open ? (
        <motion.div
          key="whats-new"
          className="fixed inset-0 z-50 grid place-items-center bg-[oklch(0_0_0/0.55)]"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          role="dialog"
          aria-modal="true"
          aria-label={t("Novedades de Toque {version}", { version: version ?? "" })}
        >
          <motion.div
            className="w-[480px] rounded-2xl bg-surface-2 p-6 shadow-dropdown"
            initial={{ opacity: 0, scale: 0.96, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.98, y: 4 }}
            transition={{ type: "spring", duration: 0.4, bounce: 0 }}
          >
            <div className="flex items-center gap-3">
              <AppIcon className="size-10" />
              <div>
                <p className="label-caps text-ink-3">{t("Novedades")}</p>
                <p className="text-[19px] font-semibold text-ink">Toque {version}</p>
              </div>
            </div>
            <ul className="mt-5 flex flex-col gap-2.5">
              {sentences(update?.notes ?? "").map((s, i) => (
                <motion.li
                  key={i}
                  className="flex gap-2.5 text-body-regular text-ink-2"
                  initial={{ opacity: 0, y: 4 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3, delay: 0.1 + i * 0.05 }}
                >
                  <span className="mt-[9px] size-1.5 shrink-0 rounded-full bg-volt" />
                  {s}
                </motion.li>
              ))}
            </ul>
            <div className="mt-6 flex justify-end">
              <Button onClick={close}>{t("Entendido")}</Button>
            </div>
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}
