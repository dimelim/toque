import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import { api, errorText, onProgress } from "@/lib/api";
import { play } from "@/lib/sound";
import type { Overview, Progress } from "@/lib/types";

export type Screen = "home" | "diagnose" | "optimize" | "tests" | "settings";

export interface Toast {
  id: number;
  tone: "ok" | "error" | "info";
  title: string;
  detail?: string;
}

interface AppState {
  screen: Screen;
  go: (s: Screen) => void;
  overview: Overview | null;
  refresh: () => Promise<void>;
  progress: Progress | null;
  task: string | null;
  run: <T>(task: string, label: string, fn: () => Promise<T>) => Promise<T | undefined>;
  toasts: Toast[];
  toast: (t: Omit<Toast, "id">) => void;
  dismiss: (id: number) => void;
  taskLabel: string;
}

const Ctx = createContext<AppState | null>(null);

export function AppProvider({ children }: { children: ReactNode }) {
  const [screen, setScreen] = useState<Screen>("home");
  const [overview, setOverview] = useState<Overview | null>(null);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [task, setTask] = useState<string | null>(null);
  const [taskLabel, setTaskLabel] = useState("");
  const [toasts, setToasts] = useState<Toast[]>([]);
  const seq = useRef(0);

  const refresh = useCallback(async () => {
    try {
      setOverview(await api.overview());
    } catch {
      return;
    }
  }, []);

  useEffect(() => {
    refresh();
    const t = setInterval(refresh, 4000);
    const off = onProgress((p) => setProgress(p.step === "done" ? null : p));
    return () => {
      clearInterval(t);
      off();
    };
  }, [refresh]);

  const toast = useCallback((t: Omit<Toast, "id">) => {
    const id = ++seq.current;
    setToasts((all) => [...all.slice(-2), { ...t, id }]);
    setTimeout(() => setToasts((all) => all.filter((x) => x.id !== id)), t.tone === "error" ? 9000 : 6000);
  }, []);

  const dismiss = useCallback((id: number) => setToasts((all) => all.filter((x) => x.id !== id)), []);

  const run = useCallback(
    async <T,>(name: string, label: string, fn: () => Promise<T>) => {
      setTask(name);
      setTaskLabel(label);
      setProgress({ task: name, step: "start", pct: 0, message: "Preparando" });
      play("start");
      try {
        const out = await fn();
        play("success");
        return out;
      } catch (e) {
        play("error");
        toast({ tone: "error", title: "No se pudo completar", detail: errorText(e) });
        return undefined;
      } finally {
        setTask(null);
        setProgress(null);
        refresh();
      }
    },
    [refresh, toast],
  );

  const value = useMemo<AppState>(
    () => ({ screen, go: setScreen, overview, refresh, progress, task, run, toasts, toast, dismiss, taskLabel }),
    [screen, overview, refresh, progress, task, run, toasts, toast, dismiss, taskLabel],
  );
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useApp() {
  const v = useContext(Ctx);
  if (!v) throw new Error("AppProvider missing");
  return v;
}
