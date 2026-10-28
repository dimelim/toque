import { useState } from "react";
import { motion } from "motion/react";
import type { ReactNode } from "react";
import { api, errorText } from "@/lib/api";
import { t } from "@/lib/i18n";
import { play } from "@/lib/sound";
import type { PrepCheck, SessionAction } from "@/lib/types";
import { useSession } from "@/hooks/use-session";
import { remember, useApp } from "@/state";
import { PageHeader } from "@/components/primitives";
import { Switch } from "@/components/ui/switch";
import { Hero } from "./session/Hero";
import { Prep } from "./session/Prep";
import { ActionCard } from "./session/Actions";
import { Notes } from "./session/Notes";

function prepCopy(id: string): { running: string; done: string; detail: string } {
  switch (id) {
    case "windowed":
      return { running: t("Cambiando Juegos en ventana"), done: t("Juegos en ventana actualizado"), detail: t("Roblox lo usará la próxima vez que lo abras.") };
    case "gamemode":
      return { running: t("Cambiando el Modo de juego de Windows"), done: t("Modo de juego de Windows actualizado"), detail: t("Windows lo usará en tu próxima partida.") };
    default:
      return { running: t("Ajustando el límite de FPS"), done: t("Límite de FPS ajustado"), detail: t("Roblox lo usará la próxima vez que lo abras.") };
  }
}

function Section({ title, hint, index, children }: { title: string; hint: string; index: number; children: ReactNode }) {
  return (
    <motion.div className="mt-8" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.3, delay: 0.05 * index, ease: [0.22, 1, 0.36, 1] }}>
      <div className="mb-3 flex items-baseline justify-between gap-4">
        <h2 className="text-[15px] font-semibold text-ink">{title}</h2>
        <p className="text-caption-1-medium text-ink-3">{hint}</p>
      </div>
      {children}
    </motion.div>
  );
}

export function Session() {
  const { run, task, toast, go, refresh } = useApp();
  const [info, setInfo] = useSession();
  const [pending, setPending] = useState("");

  const fail = (title: string, e: unknown) => {
    play("error");
    toast({ tone: "error", title, detail: errorText(e) });
  };

  const toggleMaster = async (on: boolean) => {
    play(on ? "on" : "off");
    try {
      setInfo(await api.setSession(on));
    } catch (e) {
      fail(t("No pude cambiar el Modo partida"), e);
    }
  };

  const toggleAction = async (id: string, title: string, on: boolean) => {
    setPending(id);
    try {
      setInfo(await api.setSessionAction(id, on));
      play(on ? "on" : "off");
    } catch (e) {
      fail(title, e);
    } finally {
      setPending("");
    }
  };

  const act = async (c: PrepCheck) => {
    if (c.id === "proven" || c.id === "windows") {
      remember("toque.optimizeTab", c.id === "windows" ? "windows" : "tested");
      go("optimize");
      return;
    }
    const copy = prepCopy(c.id);
    const next = await run("prep", copy.running, () => api.applyPrep(c.id));
    if (next) {
      setInfo(next);
      toast({ tone: "ok", title: copy.done, detail: copy.detail });
      refresh();
    }
  };

  const verify = async (a: SessionAction) => {
    const rec = await run("verify", t("Probando {title}", { title: a.title.toLowerCase() }), () => api.verify(a.verifyId));
    if (rec) {
      toast({ tone: rec.verdict === "improves" ? "ok" : "info", title: rec.verdict === "improves" ? t("Mejora confirmada en tu PC") : t("Sin mejora clara en tu PC"), detail: rec.summary });
      setInfo(await api.session());
    }
  };

  const on = !!info?.enabled;
  return (
    <div>
      <PageHeader eyebrow={t("Partida")} title={t("Modo partida")}>
        <Switch size="lg" isSelected={on} isDisabled={!info} onChange={toggleMaster} aria-label={t("Activar durante mis partidas")}>
          {on ? t("Encendido") : t("Apagado")}
        </Switch>
      </PageHeader>

      <Hero info={info} />

      {info && info.prep.length > 0 ? (
        <Section title={t("Antes de jugar")} hint={t("Cambios que se aplican al abrir Roblox")} index={1}>
          <Prep checks={info.prep} disabled={!!task} onAct={act} />
        </Section>
      ) : null}

      {info ? (
        <Section title={t("Durante la partida")} hint={t("Todo se deshace al cerrar Roblox")} index={2}>
          <div className="flex flex-col gap-3">
            {info.actions.map((a) => (
              <ActionCard
                key={a.id}
                action={a}
                busy={pending === a.id}
                taskActive={!!task}
                onToggle={(v) => toggleAction(a.id, a.title, v)}
                onVerify={() => verify(a)}
              />
            ))}
          </div>
        </Section>
      ) : null}

      {info ? (
        <Section title={t("Qué te está frenando")} hint={t("Solo lectura, Toque no toca nada aquí")} index={3}>
          <Notes info={info} />
        </Section>
      ) : null}
    </div>
  );
}
