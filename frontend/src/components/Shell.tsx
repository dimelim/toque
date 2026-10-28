import { useEffect, useRef } from "react";
import type { ComponentType, ReactNode } from "react";
import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import { win, api } from "@/lib/api";
import { PRIVATE, regionCity } from "@/lib/regions";
import { play } from "@/lib/sound";
import { installUpdate } from "@/lib/updates";
import { useAccount } from "@/hooks/use-account";
import { useCountUp } from "@/hooks/use-count-up";
import { useRepo } from "@/hooks/use-repo";
import { usePlay } from "@/hooks/use-play";
import { useApp } from "@/state";
import type { Screen } from "@/state";
import { Wordmark } from "./Logo";
import { WhatsNew } from "./WhatsNew";
import { ProgressBar } from "./primitives";
import { Button } from "./ui/button";
import {
  IconClose,
  IconDiagnose,
  IconHome,
  IconMaximize,
  IconMinimize,
  IconOptimize,
  IconKeyboard,
  IconPlay,
  IconSettings,
  IconTests,
  IconOk,
  IconBad,
  IconInfo,
  IconDownload,
  IconGithub,
  IconStar,
  IconRocket,
} from "./icons";

const nav: { id: Screen; label: string; icon: ComponentType<{ className?: string }> }[] = [
  { id: "home", label: "Inicio", icon: IconHome },
  { id: "session", label: "Partida", icon: IconRocket },
  { id: "diagnose", label: "Diagnóstico", icon: IconDiagnose },
  { id: "optimize", label: "Optimizar", icon: IconOptimize },
  { id: "hud", label: "HUD", icon: IconKeyboard },
  { id: "tests", label: "Pruebas", icon: IconTests },
  { id: "settings", label: "Ajustes", icon: IconSettings },
];

function WindowControls() {
  const btn = "no-drag grid h-10 w-12 place-items-center text-ink-3 outline-none transition-colors duration-150 hover:bg-surface-2 hover:text-ink focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-border-focus-ring";
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

function useWatchToast() {
  const { watch, toast } = useApp();
  const announced = useRef("");
  useEffect(() => {
    const joined = watch?.joined;
    if (!joined || announced.current === joined.jobId) return;
    announced.current = joined.jobId;
    play("success");
    toast({ tone: "ok", title: `Entraste en una partida de ${joined.city ?? regionCity(watch?.region ?? "")}`, detail: "El vigía la encontró y te metió en cuanto hubo sitio." });
  }, [watch, toast]);
  useEffect(() => {
    const note = watch?.note;
    if (!note || announced.current === note) return;
    announced.current = note;
    play("error");
    toast({ tone: "info", title: "Vigía detenido", detail: note, sticky: true });
  }, [watch, toast]);
}

function RobloxStatus() {
  const { overview, task, region, privateLink, watch } = useApp();
  const start = usePlay();
  const account = useAccount();
  const r = overview?.roblox;
  if (!r) return null;
  const inGame = r.inGame;
  return (
    <div className="rounded-xl bg-surface-1 p-3 shadow-[inset_0_0_0_1px_var(--color-line)]">
      {account ? (
        <div className="mb-3 flex items-center gap-2.5">
          {account.avatar ? <img src={account.avatar} alt="" className="size-9 shrink-0 rounded-full bg-surface-3" /> : <span className="size-9 shrink-0 rounded-full bg-surface-3" />}
          <span className="min-w-0">
            <span className="block truncate text-body-medium text-ink">{account.displayName || account.username}</span>
            <span className="block truncate text-caption-1-medium text-ink-3">@{account.username}</span>
          </span>
        </div>
      ) : null}
      <div className="flex items-center gap-2">
        <span className={cx("relative size-2 rounded-full", inGame ? "live-dot text-volt bg-volt" : r.running ? "live-dot text-warn bg-warn" : "bg-ink-4")} />
        <span className="text-caption-1-semibold text-ink">{!r.installed ? "Roblox no instalado" : inGame ? "En partida" : r.running ? "Roblox abierto" : "Roblox cerrado"}</span>
      </div>
      <p className="mt-1 truncate text-caption-1-medium text-ink-3">{inGame ? `Servidor ${r.serverIp}` : watch?.active ? `Vigilando ${regionCity(watch.region)}` : region === PRIVATE ? "Servidor privado" : region ? `Servidor en ${regionCity(region)}` : "Realistic Street Soccer"}</p>
      {overview?.session === "active" ? (
        <p className="mt-1 flex items-center gap-1.5 text-caption-1-semibold text-volt">
          <IconRocket className="size-3.5" />
          Modo partida activo
        </p>
      ) : null}
      {r.installed ? (
        <Button size="small" variant="secondary" className="mt-3 w-full" leadingIcon={IconPlay} disabled={!!task || (region === PRIVATE && !privateLink.trim())} onClick={start}>
          {r.running ? "Cambiar de servidor" : "Jugar"}
        </Button>
      ) : null}
    </div>
  );
}

function UpdateCard() {
  const { update, task, toast } = useApp();
  return (
    <AnimatePresence initial={false}>
      {update?.state === "ready" ? (
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 4 }}
          transition={{ type: "spring", duration: 0.4, bounce: 0 }}
          className="rounded-xl bg-surface-1 p-3 shadow-[inset_0_0_0_1px_var(--color-line)]"
        >
          <div className="flex items-center gap-2">
            <IconDownload className="size-4 text-volt" />
            <span className="text-caption-1-semibold text-ink">Toque {update.latest} lista</span>
          </div>
          <p className="mt-1 text-caption-1-medium text-ink-3">Se instala sola al volver a abrir Toque.</p>
          <Button
            size="small"
            className="mt-3 w-full"
            disabled={!!task}
            onClick={() => {
              play("tap");
              installUpdate(toast);
            }}
          >
            Actualizar ahora
          </Button>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
}

function RepoLink({ version }: { version: string }) {
  const repo = useRepo();
  const stars = useCountUp(repo && repo.stars >= 0 ? repo.stars : 0, 900);
  return (
    <button
      onClick={() => {
        play("tap");
        win.open(repo?.url ?? "https://github.com/dimelim/toque");
      }}
      aria-label="Ver el código de Toque en GitHub y darle una estrella"
      className="group flex h-9 items-center gap-2 rounded-[10px] px-3 text-left outline-none transition-colors duration-150 hover:bg-surface-1 focus-visible:ring-2 focus-visible:ring-border-focus-ring"
    >
      <IconGithub className="size-4 shrink-0 text-ink-3 transition-colors duration-150 group-hover:text-ink" />
      <span className="min-w-0 truncate text-caption-1-medium text-ink-3 transition-colors duration-150 group-hover:text-ink-2">Toque {version}</span>
      <AnimatePresence initial={false}>
        {repo && repo.stars >= 0 ? (
          <motion.span
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ type: "spring", duration: 0.45, bounce: 0.35 }}
            className="ml-auto flex shrink-0 items-center gap-1 rounded-full bg-surface-2 px-2 py-0.5 text-caption-1-semibold tabular-nums text-ink-2 transition-colors duration-150 group-hover:text-ink"
          >
            <IconStar className="size-3 text-volt transition-transform duration-300 ease-out group-hover:rotate-[72deg] group-hover:scale-125" />
            {stars}
          </motion.span>
        ) : null}
      </AnimatePresence>
    </button>
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
                "group relative flex h-10 items-center gap-3 rounded-[10px] px-3 text-left text-body-medium outline-none transition-[background-color,color] duration-150 focus-visible:ring-2 focus-visible:ring-border-focus-ring",
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
        <UpdateCard />
        <RobloxStatus />
        <RepoLink version={overview?.version ?? ""} />
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
  useWatchToast();
  return (
    <div className="flex h-full bg-canvas text-ink">
      <WhatsNew />
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
