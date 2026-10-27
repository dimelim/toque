import { useState } from "react";
import { motion } from "motion/react";
import * as Service from "@wails/go/app/Service";
import { errorText, win } from "@/lib/api";
import { play } from "@/lib/sound";
import { Wordmark } from "@/components/Logo";
import { ProgressBar } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { IconClose, IconMinimize } from "@/components/icons";

export interface InstallState {
  installed: boolean;
  installedVersion: string;
  dir: string;
  runningInstalled: boolean;
  version: string;
  newer: boolean;
}

const points = [
  { n: "01", t: "Mide antes de tocar nada", d: "Tu ping real dentro de Realistic Street Soccer, tus FPS y cada tirón, sin modificar el juego." },
  { n: "02", t: "Solo cambios con pruebas", d: "Cada optimización se valida alternando con y sin ella. Lo que no mejora, no se aplica." },
  { n: "03", t: "Todo se puede deshacer", d: "Toque guarda cómo estaba cada ajuste y lo restaura con un clic o al desinstalar." },
];

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
    <div className="pitch-lines relative flex h-full flex-col bg-canvas">
      <div className="drag flex h-10 shrink-0 items-start justify-end">
        <button aria-label="Minimizar" className="no-drag grid h-10 w-12 place-items-center text-ink-3 hover:bg-surface-2 hover:text-ink" onClick={() => win.minimise()}>
          <IconMinimize className="size-4" />
        </button>
        <button aria-label="Cerrar" className="no-drag grid h-10 w-12 place-items-center text-ink-3 hover:bg-bad hover:text-canvas" onClick={() => win.quit()}>
          <IconClose className="size-4" />
        </button>
      </div>
      <div className="mx-auto flex w-full max-w-[920px] flex-1 flex-col justify-center px-10 pb-16">
        <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}>
          <Wordmark className="h-[40px] text-ink" />
          <h1 className="mt-10 max-w-[18ch] text-[44px] leading-[1.02] font-semibold tracking-[-0.03em] text-ink">
            {update ? `Actualizar a la versión ${state.version}` : "Juega al toque."}
          </h1>
          <p className="mt-4 max-w-[54ch] text-[15px] leading-relaxed text-ink-2">
            {update
              ? `Tienes la versión ${state.installedVersion}. La actualización conserva tus pruebas, tus resultados y los cambios que tengas activos.`
              : "Toque mide el retraso real de Roblox en tu PC y aplica solo los cambios que demuestran reducirlo."}
          </p>
        </motion.div>
        {!update ? (
          <div className="mt-12 grid grid-cols-3 gap-6">
            {points.map((p, i) => (
              <motion.div
                key={p.n}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4, delay: 0.1 + i * 0.08, ease: [0.22, 1, 0.36, 1] }}
                className="shadow-[inset_0_1px_0_var(--color-line-strong)] pt-4"
              >
                <p className="num text-[15px] font-semibold text-volt">{p.n}</p>
                <p className="mt-2 text-body-medium text-ink">{p.t}</p>
                <p className="mt-1 text-body-regular text-ink-2">{p.d}</p>
              </motion.div>
            ))}
          </div>
        ) : null}
        <div className="mt-12 flex items-center gap-6">
          <Button disabled={busy} onClick={go} className="h-11 px-5 text-[15px]">
            {busy ? (update ? "Actualizando" : "Instalando") : update ? "Actualizar ahora" : "Instalar Toque"}
          </Button>
          {!update ? (
            <Switch size="sm" isSelected={desktop} onChange={setDesktop} isDisabled={busy}>
              <span className="text-body-regular text-ink-2">Acceso directo en el escritorio</span>
            </Switch>
          ) : null}
          <Button variant="ghost" disabled={busy} onClick={onSkip}>
            {update ? "Abrir sin actualizar" : "Probar sin instalar"}
          </Button>
        </div>
        <div className="mt-5 h-5 max-w-[420px]">
          {busy ? <ProgressBar value={0} indeterminate /> : null}
          {error ? <p className="text-body-regular text-bad">{error}</p> : null}
        </div>
        <p className="mt-6 text-caption-1-medium text-ink-4">Se instala en {state.dir}. Necesita permisos de administrador para medir con las herramientas de Windows.</p>
      </div>
    </div>
  );
}
