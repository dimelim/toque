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
  inputLatencyMs: number;
  inputEvents: number;
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
  windows: WindowPoint[] | null;
}

export interface ExperimentGroup {
  tweak: string;
  title: string;
  stress: string;
  verdict: "improves" | "unconfirmed" | "no-effect" | "worse" | "tradeoff" | string;
  pooled: Record<string, Delta>;
  tests: ExperimentRecord[];
}

export interface WindowPoint {
  arm: "A" | "B" | string;
  valid: boolean;
  values: Record<string, number>;
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
  cached: boolean;
  city?: string;
  country?: string;
  error?: string;
}

export interface PingPoint {
  at: string;
  rttMs: number;
  serverIp: string;
}

export interface Region {
  key: string;
  city: string;
  area: string;
  country: string;
  rttMs: number;
}

export interface PlayResult {
  first: ServerProbe;
  final: ServerProbe;
  hops: number;
  bestMs: number;
  region: string;
  found: boolean;
  tried: number;
  private: boolean;
  alreadyOpen: boolean;
  exitCountry: string;
}

export interface FlagEntry {
  name: string;
  value: string;
}

export interface HudConfig {
  enabled: boolean;
  x: number;
  y: number;
  scale: number;
  opacity: number;
  showKeys: boolean;
  showMouse: boolean;
  showPing: boolean;
  move: boolean;
  demo: boolean;
}

export interface HudLayout {
  screenW: number;
  screenH: number;
  panelW: number;
  panelH: number;
}

export interface Account {
  userId: number;
  username: string;
  displayName: string;
  avatar: string;
}

export interface WatchInfo {
  region: string;
  active: boolean;
  since: string;
  checked: number;
  joined?: ServerProbe;
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
  mbps: number;
}

export interface BloatResult {
  target: string;
  phases: LoadPhase[] | null;
  addedMs: number;
  maxLossPct: number;
  grade: string;
}

export interface UpdateInfo {
  current: string;
  latest: string;
  notes: string;
  state: "idle" | "checking" | "downloading" | "ready" | "current" | "error" | string;
  error: string;
  checked: string;
}

export interface Progress {
  task: string;
  step: string;
  pct: number;
  message: string;
}

export interface Repo {
  name: string;
  url: string;
  stars: number;
}

export type SessionState = "off" | "waiting" | "active" | "paused";

export interface SessionProc {
  name: string;
  cpu: number;
  ramMb: number;
}

export interface SessionAction {
  id: "self" | "windows" | "power" | "background" | "watch";
  title: string;
  what: string;
  tested: boolean;
  enabled: boolean;
  status: "off" | "waiting" | "active";
  detail: string;
  procs: SessionProc[];
  verifyId: string;
  experiment: ExperimentRecord | null;
}

export interface SessionNote {
  id: string;
  tone: "info" | "warn";
  title: string;
  detail: string;
  procs: SessionProc[];
}

export interface PrepCheck {
  id: "fps-cap" | "proven";
  title: string;
  detail: string;
  done: boolean;
  action: string;
  blocked: string;
}

export interface SessionInfo {
  enabled: boolean;
  state: SessionState;
  since: string;
  actions: SessionAction[];
  notes: SessionNote[];
  prep: PrepCheck[];
}

export interface WindowState {
  calm: boolean;
}
