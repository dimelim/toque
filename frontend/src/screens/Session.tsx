import { useState } from "react";
import { motion } from "motion/react";
import type { ReactNode } from "react";
import { api, errorText } from "@/lib/api";
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
      fail("No pude cambiar el modo partida", e);
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
    const windowed = c.id === "windowed";
    const next = await run("prep", windowed ? "Activando la optimización de juegos en ventana" : "Ajustando el límite de FPS", () => api.applyPrep(c.id));
    if (next) {
      setInfo(next);
      toast({ tone: "ok", title: windowed ? "Optimización de juegos en ventana activada" : "Límite de FPS ajustado", detail: "Roblox lo usará la próxima vez que lo abras." });
      refresh();
    }
  };

  const verify = async (a: SessionAction) => {
    const rec = await run("verify", `Probando ${a.title.toLowerCase()}`, () => api.verify(a.verifyId));
    if (rec) {
      toast({ tone: rec.verdict === "improves" ? "ok" : "info", title: rec.verdict === "improves" ? "Mejora confirmada en tu PC" : "Sin mejora clara en tu PC", detail: rec.summary });
      setInfo(await api.session());
    }
  };

  const on = !!info?.enabled;
  return (
    <div>
      <PageHeader eyebrow="Partida" title="Modo partida">
        <Switch size="lg" isSelected={on} isDisabled={!info} onChange={toggleMaster} aria-label="Activar durante mis partidas">
          {on ? "Encendido" : "Apagado"}
        </Switch>
      </PageHeader>

      <Hero info={info} />

      {info && info.prep.length > 0 ? (
        <Section title="Antes de jugar" hint="Cambios que se aplican al abrir Roblox" index={1}>
          <Prep checks={info.prep} disabled={!!task} onAct={act} />
        </Section>
      ) : null}

      {info ? (
        <Section title="Durante la partida" hint="Todo se deshace al cerrar Roblox" index={2}>
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
        <Section title="Qué te está frenando" hint="Solo lectura, Toque no toca nada aquí" index={3}>
          <Notes info={info} />
        </Section>
      ) : null}
    </div>
  );
}
