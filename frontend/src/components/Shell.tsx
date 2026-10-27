import type { ComponentType, ReactNode } from "react";
import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import { win, api } from "@/lib/api";
import { play } from "@/lib/sound";
import { useApp } from "@/state";
import type { Screen } from "@/state";
import { Wordmark } from "./Logo";
import { ProgressBar } from "./primitives";
import { Button } from "./ui/button";
import {
  IconClose,
  IconDiagnose,
  IconHome,
  IconMaximize,
  IconMinimize,
  IconOptimize,
  IconPlay,
  IconSettings,
  IconTests,
  IconOk,
  IconBad,
  IconInfo,
} from "./icons";

const nav: { id: Screen; label: string; icon: ComponentType<{ className?: string }> }[] = [
  { id: "home", label: "Inicio", icon: IconHome },
  { id: "diagnose", label: "Diagnóstico", icon: IconDiagnose },
  { id: "optimize", label: "Optimizar", icon: IconOptimize },
  { id: "tests", label: "Pruebas", icon: IconTests },
  { id: "settings", label: "Ajustes", icon: IconSettings },
];

function WindowControls() {
  const btn = "no-drag grid h-10 w-12 place-items-center text-ink-3 transition-colors duration-150 hover:bg-surface-2 hover:text-ink";
  return (
    <div className="flex">
      <button aria-label="Minimizar" className={btn} onClick={() => win.minimise()}>
        <IconMinimize className="size-4" />
      </button>
      <button aria-label="Maximizar" className={btn} onClick={() => win.toggleMaximise()}>
        <IconMaximize className="size-3.5" />
      </button>
      <button aria-label="Cerrar" className={cx(btn, "hover:bg-bad hover:text-canvas")} onClick={() => win.quit()}>
        <IconClose className="size-4" />
      </button>
    </div>
  );
}

function RobloxStatus() {
  const { overview, run, task } = useApp();
  const r = overview?.roblox;
  if (!r) return null;
  const inGame = r.inGame;
  return (
    <div className="rounded-xl bg-surface-1 p-3 shadow-[inset_0_0_0_1px_var(--color-line)]">
      <div className="flex items-center gap-2">
        <span className={cx("live-dot relative size-2 rounded-full", inGame ? "text-volt bg-volt" : r.running ? "text-warn bg-warn" : "bg-ink-4 text-transparent")} />
        <span className="text-caption-1-semibold text-ink">{!r.installed ? "Roblox no instalado" : inGame ? "En partida" : r.running ? "Roblox abierto" : "Roblox cerrado"}</span>
      </div>
      <p className="mt-1 truncate text-caption-1-medium text-ink-3">{inGame ? `Servidor ${r.serverIp}` : "Realistic Street Soccer"}</p>
      {!r.running && r.installed ? (
        <Button
          size="small"
          variant="secondary"
          className="mt-3 w-full"
          leadingIcon={IconPlay}
          disabled={!!task}
          onClick={() => {
            play("tap");
            run("play", "Abriendo Roblox", () => api.play());
          }}
        >
          Jugar
        </Button>
      ) : null}
    </div>
  );
}

function Sidebar() {
  const { screen, go, overview } = useApp();
  return (
    <aside className="flex w-[228px] shrink-0 flex-col px-3 pb-3">
      <div className="drag flex h-14 items-center px-3 pt-1">
        <Wordmark className="h-[22px] text-ink" />
      </div>
      <nav className="mt-3 flex flex-col gap-0.5" aria-label="Secciones">
        {nav.map((n) => {
          const active = screen === n.id;
          const Icon = n.icon;
          return (
            <button
              key={n.id}
              onClick={() => {
                if (!active) play("tap");
                go(n.id);
              }}
              aria-current={active ? "page" : undefined}
              className={cx(
                "group relative flex h-10 items-center gap-3 rounded-[10px] px-3 text-left text-body-medium transition-[background-color,color] duration-150",
                active ? "bg-surface-2 text-ink" : "text-ink-2 hover:bg-surface-1 hover:text-ink",
              )}
            >
              <Icon className={cx("size-[18px] shrink-0 transition-colors duration-150", active ? "text-volt" : "text-ink-3 group-hover:text-ink-2")} />
              {n.label}
            </button>
          );
        })}
      </nav>
      <div className="mt-auto flex flex-col gap-3">
        <RobloxStatus />
        <p className="px-3 text-caption-1-medium text-ink-4">Toque {overview?.version ?? ""}</p>
      </div>
    </aside>
  );
}

function TaskPanel() {
  const { task, progress, taskLabel } = useApp();
  return (
    <AnimatePresence initial={false}>
      {task ? (
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 8 }}
          transition={{ type: "spring", duration: 0.35, bounce: 0 }}
          className="pointer-events-auto w-[360px] rounded-2xl bg-surface-3 p-4 shadow-dropdown"
          role="status"
          aria-live="polite"
        >
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-body-medium text-ink">{taskLabel}</p>
              <p className="mt-0.5 truncate text-caption-1-medium text-ink-2">{progress?.message || "Trabajando"}</p>
            </div>
            <Button size="xs" variant="ghost" onClick={() => api.cancel()}>
              Cancelar
            </Button>
          </div>
          <div className="mt-3">
            <ProgressBar value={progress?.pct ?? 0} indeterminate={!progress || progress.pct <= 0.02} />
          </div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}

function Toasts() {
  const { toasts, dismiss } = useApp();
  return (
    <AnimatePresence initial={false}>
      {toasts.map((t) => {
        const Icon = t.tone === "ok" ? IconOk : t.tone === "error" ? IconBad : IconInfo;
        return (
          <motion.button
            key={t.id}
            layout
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 8 }}
            transition={{ type: "spring", duration: 0.35, bounce: 0 }}
            onClick={() => dismiss(t.id)}
            className="pointer-events-auto flex w-[360px] items-start gap-3 rounded-2xl bg-surface-3 p-4 text-left shadow-dropdown"
          >
            <Icon className={cx("mt-0.5 size-[18px] shrink-0", t.tone === "ok" ? "text-volt" : t.tone === "error" ? "text-bad" : "text-ink-2")} />
            <span className="min-w-0">
              <span className="block text-body-medium text-ink">{t.title}</span>
              {t.detail ? <span className="mt-0.5 block text-caption-1-medium text-ink-2">{t.detail}</span> : null}
            </span>
          </motion.button>
        );
      })}
    </AnimatePresence>
  );
}

export function Shell({ children }: { children: ReactNode }) {
  const { screen } = useApp();
  return (
    <div className="flex h-full bg-canvas text-ink">
      <Sidebar />
      <div className="relative flex min-w-0 flex-1 flex-col">
        <div className="drag flex h-10 shrink-0 items-start justify-end">
          <WindowControls />
        </div>
        <main className="relative min-h-0 flex-1 overflow-y-auto px-8 pb-10">
          <AnimatePresence mode="wait" initial={false}>
            <motion.div
              key={screen}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.18, ease: [0.22, 1, 0.36, 1] }}
              className="mx-auto max-w-[1040px]"
            >
              {children}
            </motion.div>
          </AnimatePresence>
        </main>
        <div className="pointer-events-none absolute right-6 bottom-6 flex flex-col items-end gap-2">
          <Toasts />
          <TaskPanel />
        </div>
      </div>
    </div>
  );
}
