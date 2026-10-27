import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import { api, errorText, onProgress, onUpdate } from "@/lib/api";
import { play } from "@/lib/sound";
import type { Overview, Progress, UpdateInfo } from "@/lib/types";

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
  update: UpdateInfo | null;
  setUpdate: (u: UpdateInfo) => void;
  region: string;
  setRegion: (key: string) => void;
}

const Ctx = createContext<AppState | null>(null);

const regionStore = "toque.region";

function savedRegion() {
  try {
    return localStorage.getItem(regionStore) ?? "";
  } catch {
    return "";
  }
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [screen, setScreen] = useState<Screen>("home");
  const [overview, setOverview] = useState<Overview | null>(null);
  const [progress, setProgress] = useState<Progress | null>(null);
  const [task, setTask] = useState<string | null>(null);
  const [taskLabel, setTaskLabel] = useState("");
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [update, setUpdate] = useState<UpdateInfo | null>(null);
  const [region, setRegionState] = useState(savedRegion);
  const seq = useRef(0);

  const setRegion = useCallback((key: string) => {
    setRegionState(key);
    try {
      localStorage.setItem(regionStore, key);
    } catch {
      return;
    }
  }, []);

  const refresh = useCallback(async () => {
    try {
      setOverview(await api.overview());
    } catch {
      return;
    }
  }, []);

  useEffect(() => {
    refresh();
    api.updateStatus().then(setUpdate).catch(() => undefined);
    const t = setInterval(refresh, 4000);
    const off = onProgress((p) => setProgress(p.step === "done" ? null : p));
    const offUpdate = onUpdate(setUpdate);
    return () => {
      clearInterval(t);
      off();
      offUpdate();
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
    () => ({ screen, go: setScreen, overview, refresh, progress, task, run, toasts, toast, dismiss, taskLabel, update, setUpdate, region, setRegion }),
    [screen, overview, refresh, progress, task, run, toasts, toast, dismiss, taskLabel, update, region, setRegion],
  );
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useApp() {
  const v = useContext(Ctx);
  if (!v) throw new Error("AppProvider missing");
  return v;
}
