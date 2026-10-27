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
  if (r.needVpn) {
    return {
      title: `Tu conexión sale por ${countryName(r.exitCountry)}`,
      detail: `Enciende una VPN de ${countryName(regionCountry(r.region))} y pulsa Jugar otra vez.`,
    };
  }
  if (r.region && r.found) {
    return {
      title: `Tu servidor privado se abrió en ${f.city}`,
      detail: "Ya puedes apagar la VPN. Si Roblox te desconecta, pulsa Jugar enseguida para volver al mismo servidor mientras siga abierto.",
    };
  }
  if (r.region) {
    return {
      title: `Tu servidor privado se abrió${where(f)}`,
      detail: `Roblox no lo abrió en ${regionCity(r.region)}. Si ya había alguien dentro, espera a que se vacíe y vuelve a probar.`,
    };
  }
  return { title: `Dentro de tu servidor privado${where(f)}${ping(f)}` };
}

export function playMessage(r: PlayResult) {
  const f = r.final;
  if (r.private) return privateMessage(r);
  if (r.region && r.found) {
    return { title: `Dentro de un servidor${where(f)}${ping(f)}`, detail: r.tried > 1 ? `Lo encontré en el intento ${r.tried}.` : undefined };
  }
  if (r.region) {
    return { title: `No encontré servidores libres en ${regionCity(r.region)}`, detail: `Probé ${r.tried} y ahora estás en ${f.city ? `uno de ${f.city}` : "otro"}${ping(f)}.` };
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
    if (region === PRIVATE) {
      const r = await run("play", "Entrando en tu servidor privado", () => api.playPrivate(privateLink, privateRegion));
      if (r) toast({ tone: r.needVpn || (r.region && !r.found) ? "info" : "ok", ...playMessage(r) });
      return r;
    }
    const label = region ? `Buscando un servidor en ${regionCity(region)}` : "Entrando en un servidor cercano";
    const r = await run("play", label, () => api.play(region));
    if (r) toast({ tone: r.region && !r.found ? "info" : "ok", ...playMessage(r) });
    return r;
  };
}
