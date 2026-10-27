type Cue = "tap" | "on" | "off" | "success" | "error" | "start";

let ctx: AudioContext | null = null;
let enabled = readEnabled();

function readEnabled() {
  try {
    return localStorage.getItem("toque.sound") !== "off";
  } catch {
    return true;
  }
}

export function soundEnabled() {
  return enabled;
}

export function setSoundEnabled(v: boolean) {
  enabled = v;
  try {
    localStorage.setItem("toque.sound", v ? "on" : "off");
  } catch {
    return;
  }
}

function audio() {
  if (!ctx) ctx = new AudioContext();
  if (ctx.state === "suspended") ctx.resume();
  return ctx;
}

function tone(freq: number, start: number, dur: number, gain: number, type: OscillatorType = "sine", to?: number) {
  const a = audio();
  const osc = a.createOscillator();
  const g = a.createGain();
  osc.type = type;
  osc.frequency.setValueAtTime(freq, a.currentTime + start);
  if (to) osc.frequency.exponentialRampToValueAtTime(to, a.currentTime + start + dur);
  g.gain.setValueAtTime(0.0001, a.currentTime + start);
  g.gain.exponentialRampToValueAtTime(gain, a.currentTime + start + 0.006);
  g.gain.exponentialRampToValueAtTime(0.0001, a.currentTime + start + dur);
  osc.connect(g).connect(a.destination);
  osc.start(a.currentTime + start);
  osc.stop(a.currentTime + start + dur + 0.02);
}

export function play(cue: Cue) {
  if (!enabled) return;
  try {
    switch (cue) {
      case "tap":
        tone(1850, 0, 0.035, 0.035, "triangle");
        break;
      case "on":
        tone(920, 0, 0.05, 0.05, "triangle");
        tone(1380, 0.045, 0.07, 0.045, "triangle");
        break;
      case "off":
        tone(1100, 0, 0.05, 0.04, "triangle");
        tone(740, 0.045, 0.07, 0.035, "triangle");
        break;
      case "start":
        tone(520, 0, 0.12, 0.04, "sine", 780);
        break;
      case "success":
        tone(659.25, 0, 0.14, 0.05, "triangle");
        tone(987.77, 0.09, 0.22, 0.05, "triangle");
        tone(1318.5, 0.18, 0.3, 0.03, "sine");
        break;
      case "error":
        tone(220, 0, 0.18, 0.06, "sine", 140);
        break;
    }
  } catch {
    return;
  }
}
