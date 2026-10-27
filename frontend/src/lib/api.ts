import * as Service from "@wails/go/app/Service";
import { EventsOn, WindowMinimise, WindowToggleMaximise, Quit, BrowserOpenURL } from "@wails/runtime/runtime";
import type {
  BloatResult,
  Diagnosis,
  ExperimentRecord,
  Optimization,
  Overview,
  Progress,
  RunRecord,
  ServerProbe,
  ServerSearch,
  UpdateInfo,
} from "./types";

export const api = {
  overview: () => Service.Overview() as Promise<Overview>,
  diagnose: (withGame: boolean) => Service.Diagnose(withGame) as Promise<Diagnosis>,
  optimizations: () => Service.Optimizations() as Promise<Optimization[]>,
  apply: (id: string) => Service.Apply(id) as Promise<Optimization>,
  revert: (id: string) => Service.Revert(id) as Promise<Optimization>,
  applyProven: () => Service.ApplyProven() as Promise<string[]>,
  revertAll: () => Service.RevertAll() as Promise<void>,
  verify: (id: string) => Service.Verify(id) as Promise<ExperimentRecord>,
  experiments: () => Service.Experiments() as Promise<ExperimentRecord[]>,
  benchmark: (label: string) => Service.Benchmark(label) as Promise<RunRecord>,
  runs: () => Service.Runs() as Promise<RunRecord[]>,
  removeRun: (id: string) => Service.RemoveRun(id) as Promise<void>,
  findServer: (tries: number) => Service.FindServer(tries) as Promise<ServerSearch>,
  bufferbloat: () => Service.Bufferbloat() as Promise<BloatResult>,
  quickPing: () => Service.QuickPing() as Promise<ServerProbe>,
  play: () => Service.Play() as Promise<void>,
  cancel: () => Service.Cancel() as Promise<void>,
  updateStatus: () => Service.UpdateStatus() as Promise<UpdateInfo>,
  checkUpdate: () => Service.CheckUpdate() as Promise<UpdateInfo>,
  applyUpdate: () => Service.ApplyUpdate() as Promise<void>,
};

export function onProgress(cb: (p: Progress) => void) {
  return EventsOn("progress", cb);
}

export function onUpdate(cb: (u: UpdateInfo) => void) {
  return EventsOn("update", cb);
}

export const win = {
  minimise: () => WindowMinimise(),
  toggleMaximise: () => WindowToggleMaximise(),
  quit: () => Quit(),
  open: (url: string) => BrowserOpenURL(url),
};

export function errorText(e: unknown): string {
  if (typeof e === "string") return e;
  if (e instanceof Error) return e.message;
  return "Algo salió mal";
}
