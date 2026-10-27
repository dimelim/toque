import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { PRIVATE, regionCity } from "@/lib/regions";
import { play } from "@/lib/sound";
import type { PlayResult, ServerProbe } from "@/lib/types";
import { useApp } from "@/state";

function where(p: ServerProbe) {
  return p.city ? ` en ${p.city}` : "";
}

function ping(p: ServerProbe) {
  return p.rttMs ? `, a ${num(p.rttMs, 0)} ms` : "";
}

export function playMessage(r: PlayResult) {
  const f = r.final;
  if (r.private) return { title: `Dentro de tu servidor privado${where(f)}${ping(f)}` };
  if (r.region && r.found) return { title: `Dentro de una partida${where(f)}${ping(f)}` };
  if (r.region) {
    return {
      title: `No hay partidas de ${regionCity(r.region)} ahora`,
      detail: r.tried
        ? `Las ${r.tried} que encontré se llenaron antes de entrar. Pulsa Vigilar y Toque entrará en cuanto haya sitio.`
        : "RoValra no conoce ninguna con sitio. Pulsa Vigilar y Toque entrará en cuanto aparezca una.",
    };
  }
  if (!f.rttMs) return { title: f.city ? `Dentro de un servidor en ${f.city}` : "Dentro de Realistic Street Soccer" };
  if (f.jobId !== r.first.jobId) return { title: `Te moví a un servidor${where(f)}${ping(f)}`, detail: `Roblox te había puesto en uno${where(r.first)} a ${num(r.first.rttMs, 0)} ms.` };
  if (r.hops > 0) return { title: `Servidor${where(f)}${ping(f)}`, detail: `Probé ${r.hops} más y ninguno estaba más cerca.` };
  return { title: `Servidor cercano${where(f)}${ping(f)}` };
}

export function usePlay() {
  const { run, toast, region, privateLink } = useApp();
  return async () => {
    play("tap");
    if (region === PRIVATE) {
      const r = await run("play", "Entrando en tu servidor privado", () => api.playPrivate(privateLink));
      if (r) toast({ tone: "ok", ...playMessage(r) });
      return r;
    }
    const label = region ? `Buscando partidas de ${regionCity(region)}` : "Entrando en un servidor cercano";
    const r = await run("play", label, () => api.play(region));
    if (r) toast({ tone: r.region && !r.found ? "info" : "ok", ...playMessage(r) });
    return r;
  };
}
