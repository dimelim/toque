import * as Service from "@wails/go/app/Service";
import { ClipboardSetText, EventsOn, WindowMinimise, WindowToggleMaximise, Quit, BrowserOpenURL } from "@wails/runtime/runtime";
import type {
  Account,
  BloatResult,
  Diagnosis,
  ExperimentGroup,
  ExperimentRecord,
  FlagEntry,
  HudConfig,
  HudLayout,
  Optimization,
  Overview,
  PingPoint,
  PlayResult,
  Progress,
  Region,
  Repo,
  RunRecord,
  ServerProbe,
  ServerSearch,
  SessionInfo,
  UpdateInfo,
  WatchInfo,
  WindowState,
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
  experimentGroups: () => Service.ExperimentGroups() as Promise<ExperimentGroup[]>,
  benchmark: (label: string) => Service.Benchmark(label) as Promise<RunRecord>,
  runs: () => Service.Runs() as Promise<RunRecord[]>,
  removeRun: (id: string) => Service.RemoveRun(id) as Promise<void>,
  findServer: (tries: number) => Service.FindServer(tries) as Promise<ServerSearch>,
  bufferbloat: () => Service.Bufferbloat() as Promise<BloatResult>,
  quickPing: () => Service.QuickPing() as Promise<ServerProbe>,
  play: (region: string) => Service.PlayIn(region) as Promise<PlayResult>,
  playPrivate: (link: string, region: string) => Service.PlayPrivate(link, region) as Promise<PlayResult>,
  watchRegion: (region: string) => Service.WatchRegion(region) as Promise<WatchInfo>,
  stopWatch: () => Service.StopWatch() as Promise<WatchInfo>,
  watchStatus: () => Service.WatchStatus() as Promise<WatchInfo>,
  account: () => Service.Account() as Promise<Account>,
  flags: () => Service.Flags() as Promise<FlagEntry[]>,
  updateFlags: (set: Record<string, string>, clear: string[]) => Service.UpdateFlags(set, clear) as Promise<FlagEntry[]>,
  copyText: (text: string) => ClipboardSetText(text),
  hudConfig: () => Service.HudConfig() as Promise<HudConfig>,
  setHudConfig: (cfg: HudConfig) => Service.SetHudConfig(cfg) as Promise<HudConfig>,
  moveHud: (corner: string) => Service.MoveHudTo(corner) as Promise<HudConfig>,
  hudLayout: () => Service.HudLayout() as Promise<HudLayout>,
  placeHud: (x: number, y: number, save: boolean) => Service.PlaceHud(x, y, save) as Promise<HudConfig>,
  previewHud: (on: boolean) => Service.PreviewHud(on) as Promise<void>,
  cancel: () => Service.Cancel() as Promise<void>,
  noteError: (text: string) => Service.NoteError(text) as Promise<void>,
  exportReport: () => Service.ExportReport() as Promise<string>,
  repo: () => Service.Repo() as Promise<Repo>,
  session: () => Service.Session() as Promise<SessionInfo>,
  setSession: (enabled: boolean) => Service.SetSession(enabled) as Promise<SessionInfo>,
  setSessionAction: (id: string, on: boolean) => Service.SetSessionAction(id, on) as Promise<SessionInfo>,
  applyPrep: (id: string) => Service.ApplyPrep(id) as Promise<SessionInfo>,
  window: () => Service.Window() as Promise<WindowState>,
  lastDiagnosis: () => Service.LastDiagnosis() as Promise<Diagnosis | null>,
  lastBloat: () => Service.LastBloat() as Promise<BloatResult | null>,
  pingHistory: () => Service.PingHistory() as Promise<PingPoint[]>,
  regions: () => Service.Regions() as Promise<Region[]>,
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

export function onWatch(cb: (w: WatchInfo) => void) {
  return EventsOn("watch", cb);
}

export function onOverview(cb: (o: Overview) => void) {
  return EventsOn("overview", cb);
}

export function onSession(cb: (s: SessionInfo) => void) {
  return EventsOn("session", cb);
}

export function onWindow(cb: (w: WindowState) => void) {
  return EventsOn("window", cb);
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
