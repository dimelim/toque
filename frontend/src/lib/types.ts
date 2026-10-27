export interface SystemInfo {
  cpu: string;
  threads: number;
  gpu: string;
  ramGb: number;
  refreshHz: number;
  os: string;
}

export interface RobloxState {
  installed: boolean;
  version: string;
  running: boolean;
  inGame: boolean;
  placeId: number;
  jobId: string;
  serverIp: string;
}

export interface FrameSummary {
  frames: number;
  seconds: number;
  avgFps: number;
  low1Fps: number;
  low01Fps: number;
  frametimeP50: number;
  frametimeP99: number;
  frametimeStdev: number;
  stutters: number;
  stutterPerMin: number;
  renderLatencyMs: number;
  renderLatencyP99Ms: number;
  cpuBusyMs: number;
  gpuBusyMs: number;
  bound: string;
  presentMode: string;
  tearing: boolean;
}

export interface NetSummary {
  packets: number;
  seconds: number;
  rttMs: number;
  rttP95Ms: number;
  jitterMs: number;
  lossDownPct: number;
  maxGapMs: number;
  stalls100: number;
  kbpsIn: number;
  kbpsOut: number;
  rttSeries: number[] | null;
  confident: boolean;
}

export interface Disk {
  name: string;
  readsPerSec: number;
  readLatencyMs: number;
  readLatencyMaxMs: number;
}

export interface SysSummary {
  samples: number;
  cpuTotalPct: number;
  cpuRobloxPct: number;
  gpuTotalPct: number;
  gpuRobloxPct: number;
  ramAvailMb: number;
  ramAvailMinMb: number;
  ramRobloxMb: number;
  hardFaultsPerSec: number;
  hardFaultsMaxPerSec: number;
  disks: Disk[] | null;
}

export interface Window {
  label: string;
  start: string;
  seconds: number;
  jobId: string;
  serverIp: string;
  foreground: boolean;
  frames: FrameSummary;
  net: NetSummary;
  sys: SysSummary;
  errors?: string[] | null;
}

export interface RunRecord {
  id: string;
  label: string;
  at: string;
  serverIp: string;
  jobId: string;
  joinMs: number;
  applied: string[] | null;
  window: Window;
}

export interface Delta {
  a: number;
  b: number;
  pct: number;
  low: number;
  high: number;
  p: number;
  conclusive: boolean;
}

export interface ExperimentRecord {
  id: string;
  tweak: string;
  title: string;
  stress: string;
  primary: string[] | null;
  at: string;
  verdict: "improves" | "worse" | "no-effect" | "tradeoff" | string;
  summary: string;
  valid: number;
  compare: Record<string, Delta>;
}

export interface Optimization {
  id: string;
  title: string;
  category: "memoria" | "sistema" | "fps" | "red" | string;
  what: string;
  why: string;
  metric: string;
  risk: string;
  revert: string;
  verdict: "proven" | "rejected" | "pending" | "situational" | string;
  lab: string;
  kind: "live" | "relaunch" | "reboot" | string;
  applicable: boolean;
  applied: boolean;
  managed: boolean;
  detail: string;
  mine: ExperimentRecord | null;
}

export interface Overview {
  version: string;
  elevated: boolean;
  roblox: RobloxState;
  system: SystemInfo;
  applied: number;
  proven: number;
  tested: number;
  lastRun: RunRecord | null;
  busy: string;
  uptimeMin: number;
}

export interface Finding {
  id: string;
  level: "ok" | "info" | "warn" | "bad";
  area: string;
  title: string;
  detail: string;
  fix?: string;
}

export interface PathStep {
  label: string;
  addr: string;
  rttMs: number;
  jitterMs: number;
  lossPct: number;
  kind: "local" | "isp" | "game" | string;
}

export interface ProcMem {
  name: string;
  mb: number;
  count: number;
}

export interface PagefileInfo {
  drives: string[] | null;
  onHdd: boolean;
  automatic: boolean;
  systemDrive: string;
  systemIsSsd: boolean;
  systemFreeGb: number;
}

export interface Diagnosis {
  at: string;
  system: SystemInfo;
  path: PathStep[] | null;
  game: Window | null;
  pagefile: PagefileInfo;
  powerPlan: string;
  powerName: string;
  ramAvailMb: number;
  hogs: ProcMem[] | null;
  framerateCap: number;
  findings: Finding[] | null;
}

export interface ServerProbe {
  jobId: string;
  ip: string;
  rttMs: number;
  jitterMs: number;
  lossPct: number;
  playing: number;
  max: number;
  current: boolean;
  error?: string;
}

export interface ServerSearch {
  probes: ServerProbe[] | null;
  best: ServerProbe | null;
  joined: boolean;
}

export interface LoadPhase {
  phase: string;
  rttMs: number;
  p95Ms: number;
  jitterMs: number;
  lossPct: number;
}

export interface BloatResult {
  target: string;
  phases: LoadPhase[] | null;
  addedMs: number;
  grade: string;
}

export interface Progress {
  task: string;
  step: string;
  pct: number;
  message: string;
}
