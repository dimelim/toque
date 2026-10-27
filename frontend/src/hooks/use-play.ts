import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { PRIVATE, countryName, regionCity, regionCountry } from "@/lib/regions";
import { play } from "@/lib/sound";
import type { PlayResult, ServerProbe } from "@/lib/types";
import { useApp } from "@/state";

function where(p: ServerProbe) {
  return p.city ? ` en ${p.city}` : "";
}

function ping(p: ServerProbe) {
  return p.rttMs ? `, a ${num(p.rttMs, 0)} ms` : "";
}

function privateMessage(r: PlayResult) {
  const f = r.final;
  if (!r.region) return { title: `Dentro de tu servidor privado${where(f)}${ping(f)}` };
  const city = regionCity(r.region);
  if (r.found) return { title: `Dentro de tu servidor privado en ${city}${ping(f)}`, detail: `Seguirá en ${city} mientras quede alguien dentro.` };
  if (r.alreadyOpen) {
    return {
      title: `Tu servidor privado ya estaba abierto${where(f)}`,
      detail: "Roblox solo elige la ciudad cuando lo abre. Cuando no quede nadie dentro, espera un minuto a que se cierre y vuelve a pulsar Jugar.",
    };
  }
  const code = regionCountry(r.region);
  const country = countryName(code);
  const title = `Roblox abrió tu servidor${where(f)}, no en ${city}`;
  if (r.exitCountry === code) {
    return { title, detail: `Tu conexión ya sale por ${country}, pero Roblox también reparte según el sitio libre que tiene. Sal, espera un minuto y vuelve a probar.` };
  }
  const exit = r.exitCountry ? ` y tu conexión sale por ${countryName(r.exitCountry)}` : "";
  return { title, detail: `Lo abre cerca de quien entra primero${exit}. Para que se abra en ${city} tiene que entrar primero alguien desde ${country}, o tú con una VPN de ${country}.` };
}

export function playMessage(r: PlayResult) {
  const f = r.final;
  if (r.private) return privateMessage(r);
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
  const { run, toast, region, privateLink, privateRegion } = useApp();
  return async () => {
    play("tap");
    const r =
      region === PRIVATE
        ? await run("play", "Entrando en tu servidor privado", () => api.playPrivate(privateLink, privateRegion))
        : await run("play", region ? `Buscando partidas de ${regionCity(region)}` : "Entrando en un servidor cercano", () => api.play(region));
    if (r) toast({ tone: r.region && !r.found ? "info" : "ok", sticky: r.private && !!r.region && !r.found, ...playMessage(r) });
    return r;
  };
}
