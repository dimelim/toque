import { useState } from "react";
import type { ComponentType, ReactNode } from "react";
import { motion } from "motion/react";
import * as Service from "@wails/go/app/Service";
import { errorText, win } from "@/lib/api";
import { t } from "@/lib/i18n";
import { play } from "@/lib/sound";
import { SymbolReveal, Wordmark } from "@/components/Logo";
import { IconTile } from "@/components/IconTile";
import type { TileTone } from "@/components/IconTile";
import { ProgressBar } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { IconArrowRight, IconClose, IconFlask, IconGlobe, IconMinimize, IconRevert } from "@/components/icons";

export interface InstallState {
  installed: boolean;
  installedVersion: string;
  dir: string;
  runningInstalled: boolean;
  version: string;
  newer: boolean;
}

const ease = [0.22, 1, 0.36, 1] as const;

const points = (): { icon: ComponentType<{ className?: string }>; tone: TileTone; title: string; text: string }[] => [
  { icon: IconGlobe, tone: "cyan", title: t("Elige la ciudad"), text: t("Los 35 centros de datos de Roblox, también Lima, Santiago, Buenos Aires y Querétaro.") },
  { icon: IconFlask, tone: "lime", title: t("Solo cambios con pruebas"), text: t("Cada ajuste se prueba con y sin él. Lo que no mejora, no se aplica.") },
  { icon: IconRevert, tone: "amber", title: t("Todo se puede deshacer"), text: t("Toque guarda cómo estaba cada ajuste y lo restaura con un clic.") },
];

function Reveal({ delay, className, children }: { delay: number; className?: string; children: ReactNode }) {
  return (
    <motion.div className={className} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5, delay, ease }}>
      {children}
    </motion.div>
  );
}

export function Install({ state, onSkip }: { state: InstallState; onSkip: () => void }) {
  const [desktop, setDesktop] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const update = state.installed;
  const go = async () => {
    setBusy(true);
    setError("");
    play("start");
    try {
      await Service.Install(desktop);
      play("success");
    } catch (e) {
      play("error");
      setError(errorText(e));
      setBusy(false);
    }
  };
  return (
    <div className="flex h-full flex-col bg-canvas">
      <div className="drag flex h-10 shrink-0 items-start justify-end">
        <button aria-label={t("Minimizar")} className="no-drag grid h-10 w-12 place-items-center text-ink-3 outline-none hover:bg-surface-2 hover:text-ink focus-visible:ring-2 focus-visible:ring-border-focus-ring focus-visible:ring-inset" onClick={() => win.minimise()}>
          <IconMinimize className="size-4" />
        </button>
        <button aria-label={t("Cerrar")} className="no-drag grid h-10 w-12 place-items-center text-ink-3 outline-none hover:bg-bad hover:text-canvas focus-visible:ring-2 focus-visible:ring-border-focus-ring focus-visible:ring-inset" onClick={() => win.quit()}>
          <IconClose className="size-4" />
        </button>
      </div>
      <div className="mx-auto flex w-full max-w-[1040px] flex-1 items-center gap-12 px-12 pb-12">
        <div className="min-w-0 flex-1">
          <Reveal delay={0}>
            <Wordmark className="h-[30px] text-ink" />
          </Reveal>
          {!update ? (
            <Reveal delay={0.08} className="mt-8">
              <span className="inline-flex items-center gap-2 rounded-full bg-volt-soft px-3 py-1 text-caption-1-semibold text-volt">
                <span className="size-1.5 rounded-full bg-volt" />
                {t("Hecho para Realistic Street Soccer")}
              </span>
            </Reveal>
          ) : null}
          <Reveal delay={0.14} className={update ? "mt-10" : "mt-4"}>
            <h1 className="max-w-[16ch] text-[50px] leading-[1.02] font-semibold tracking-[-0.035em] text-ink">{update ? t("Actualizar a la versión {version}", { version: state.version }) : t("Juega al toque.")}</h1>
            <p className="mt-4 max-w-[46ch] text-[15px] leading-relaxed text-ink-2">
              {update
                ? t("Tienes la versión {version}. La actualización conserva tus pruebas, tus resultados y los cambios que tengas activos.", { version: state.installedVersion })
                : t("Toque mide el retraso real de Roblox en tu PC y aplica solo los cambios que demuestran reducirlo.")}
            </p>
          </Reveal>
          {!update ? (
            <div className="mt-8 flex flex-col gap-4">
              {points().map((p, i) => (
                <Reveal key={p.title} delay={0.24 + i * 0.08} className="flex items-start gap-3.5">
                  <IconTile icon={p.icon} tone={p.tone} className="size-9" />
                  <div className="min-w-0">
                    <p className="text-body-medium text-ink">{p.title}</p>
                    <p className="mt-0.5 text-body-regular text-ink-3">{p.text}</p>
                  </div>
                </Reveal>
              ))}
            </div>
          ) : null}
          <Reveal delay={update ? 0.24 : 0.5} className="mt-9 flex flex-wrap items-center gap-x-5 gap-y-3">
            <Button disabled={busy} onClick={go} trailingIcon={busy ? undefined : IconArrowRight} className="h-11 px-5 text-[15px]">
              {busy ? (update ? t("Actualizando") : t("Instalando")) : update ? t("Actualizar ahora") : t("Instalar Toque")}
            </Button>
            <Button variant="ghost" disabled={busy} onClick={onSkip}>
              {update ? t("Abrir sin actualizar") : t("Probar sin instalar")}
            </Button>
            {!update ? (
              <Switch size="sm" isSelected={desktop} onChange={setDesktop} isDisabled={busy}>
                <span className="text-body-regular text-ink-2">{t("Acceso directo en el escritorio")}</span>
              </Switch>
            ) : null}
          </Reveal>
          <div className="mt-4 h-5 max-w-[420px]">
            {busy ? <ProgressBar value={0} indeterminate /> : null}
            {error ? <p className="text-body-regular text-bad">{error}</p> : null}
          </div>
          <Reveal delay={update ? 0.3 : 0.58}>
            <p className="mt-3 max-w-[56ch] text-caption-1-medium text-ink-4">{t("Se instala en {dir}. Necesita permisos de administrador para medir con las herramientas de Windows.", { dir: state.dir })}</p>
          </Reveal>
        </div>
        <div className="flex w-[360px] shrink-0 justify-center" aria-hidden>
          <SymbolReveal className="h-[300px]" />
        </div>
      </div>
    </div>
  );
}
