import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { t } from "@/lib/i18n";
import { PRIVATE, countryName, regionCity, regionCountry } from "@/lib/regions";
import { play } from "@/lib/sound";
import type { PlayResult, ServerProbe } from "@/lib/types";
import { useApp } from "@/state";

function where(p: ServerProbe) {
  return p.city ? ` ${t("en {city}", { city: p.city })}` : "";
}

function ping(p: ServerProbe) {
  return p.rttMs ? `, ${t("a {ms} ms", { ms: num(p.rttMs, 0) })}` : "";
}

function privateMessage(r: PlayResult) {
  const f = r.final;
  if (!r.region) return { title: t("Dentro de tu servidor privado{where}{ping}", { where: where(f), ping: ping(f) }) };
  const city = regionCity(r.region);
  if (r.found) return { title: t("Dentro de tu servidor privado en {city}{ping}", { city, ping: ping(f) }), detail: t("Seguirá en {city} mientras quede alguien dentro.", { city }) };
  if (r.alreadyOpen) {
    return {
      title: t("Tu servidor privado ya estaba abierto{where}", { where: where(f) }),
      detail: t("Roblox solo elige la ciudad cuando lo abre. Cuando no quede nadie dentro, espera un minuto a que se cierre y vuelve a pulsar Jugar."),
    };
  }
  const code = regionCountry(r.region);
  const country = countryName(code);
  const title = t("Roblox abrió tu servidor{where}, no en {city}", { where: where(f), city });
  if (r.exitCountry === code) {
    return { title, detail: t("Tu conexión ya sale por {country}, pero Roblox también reparte según el sitio libre que tiene. Sal, espera un minuto y vuelve a probar.", { country }) };
  }
  const exit = r.exitCountry ? ` ${t("y tu conexión sale por {country}", { country: countryName(r.exitCountry) })}` : "";
  return { title, detail: t("Lo abre cerca de quien entra primero{exit}. Para que se abra en {city} tiene que entrar primero alguien desde {country}, o tú con una VPN de {country}.", { exit, city, country }) };
}

export function playMessage(r: PlayResult) {
  const f = r.final;
  if (r.private) return privateMessage(r);
  if (r.region && r.found) return { title: t("Dentro de una partida{where}{ping}", { where: where(f), ping: ping(f) }) };
  if (r.region) {
    return {
      title: t("No hay partidas de {city} ahora", { city: regionCity(r.region) }),
      detail: r.tried
        ? t("Las {n} que encontré se llenaron antes de entrar. Pulsa Vigilar y Toque entrará en cuanto haya sitio.", { n: r.tried })
        : t("RoValra no conoce ninguna con sitio. Pulsa Vigilar y Toque entrará en cuanto aparezca una."),
    };
  }
  if (!f.rttMs) return { title: f.city ? t("Dentro de un servidor en {city}", { city: f.city }) : t("Dentro de Realistic Street Soccer") };
  if (f.jobId !== r.first.jobId) {
    return {
      title: t("Te moví a un servidor{where}{ping}", { where: where(f), ping: ping(f) }),
      detail: t("Roblox te había puesto en uno{where} a {ms} ms.", { where: where(r.first), ms: num(r.first.rttMs, 0) }),
    };
  }
  if (r.hops > 0) return { title: t("Servidor{where}{ping}", { where: where(f), ping: ping(f) }), detail: t("Probé {n} más y ninguno estaba más cerca.", { n: r.hops }) };
  return { title: t("Servidor cercano{where}{ping}", { where: where(f), ping: ping(f) }) };
}

export function usePlay() {
  const { run, toast, region, privateLink, privateRegion } = useApp();
  return async () => {
    play("tap");
    const r =
      region === PRIVATE
        ? await run("play", t("Entrando en tu servidor privado"), () => api.playPrivate(privateLink, privateRegion))
        : await run("play", region ? t("Buscando partidas de {city}", { city: regionCity(region) }) : t("Entrando en un servidor cercano"), () => api.play(region));
    if (r) toast({ tone: r.region && !r.found ? "info" : "ok", sticky: r.private && !!r.region && !r.found, ...playMessage(r) });
    return r;
  };
}
