import { useState } from "react";
import { Input, Label, TextField } from "react-aria-components";
import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { PRIVATE, bestRegion, countryName, regionCity, regionCountry } from "@/lib/regions";
import type { Region, ServerProbe } from "@/lib/types";
import { usePlay } from "@/hooks/use-play";
import { useApp } from "@/state";
import { cx } from "@/utils/cx";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { StatusDot } from "@/components/ui/status-dot";
import { RegionSelect } from "@/components/RegionSelect";
import { IconPlay, IconSignal } from "@/components/icons";

function Stat({ label, value, tone }: { label: string; value: number | null | undefined; tone: "volt" | "ink" }) {
  return (
    <div className="rounded-lg px-3 py-2">
      <p className="text-caption-1-medium text-ink-3">{label}</p>
      <p className={cx("num mt-0.5 text-[26px] leading-none font-semibold", tone === "volt" ? "text-volt" : "text-ink")}>
        {num(value, 0)}
        <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
      </p>
    </div>
  );
}

function intro(inGame: boolean, region: string, privateRegion: string) {
  if (inGame) return { title: "Ya estás jugando", text: "Mide el ping del servidor en el que estás, escuchando el tráfico real del juego durante 8 segundos." };
  if (region === PRIVATE && privateRegion) {
    return {
      title: `Abrir tu servidor privado en ${regionCity(privateRegion)}`,
      text: `Vacíalo, enciende una VPN de ${countryName(regionCountry(privateRegion))} y pulsa Jugar. Toque comprueba la ciudad en la que se abre y te dice cuándo apagar la VPN.`,
    };
  }
  if (region === PRIVATE) return { title: "Jugar en tu servidor privado", text: "Pega el enlace de invitación. Toque lo guarda solo en este PC y entra directo." };
  if (region) {
    const city = regionCity(region);
    return { title: `Jugar en ${city}`, text: `Toque prueba servidores públicos y descarta los que no son de ${city} hasta dar con uno. Suele tardar entre medio minuto y dos.` };
  }
  return { title: "Jugar en un servidor cercano", text: "Roblox te manda a un servidor lejano en una de cada seis partidas. Toque lo comprueba al entrar y, si pasa, te mueve a uno cercano." };
}

export function PlayCard({ regions, onDone }: { regions: Region[]; onDone: () => void }) {
  const { overview, run, task, region, setRegion, privateLink, setPrivateLink, privateRegion, setPrivateRegion } = useApp();
  const start = usePlay();
  const [probe, setProbe] = useState<ServerProbe | null>(null);
  const r = overview?.roblox;
  const best = bestRegion(regions);
  const fresh = probe && r?.inGame && probe.jobId === r.jobId ? probe : null;
  const status = !r?.installed ? "Roblox no está instalado" : r.inGame ? "En partida" : r.running ? "Roblox abierto" : "Roblox cerrado";
  const text = intro(!!r?.inGame, region, privateRegion);
  const needsLink = region === PRIVATE && !privateLink.trim();

  const measure = async () => {
    const p = await run("ping", "Midiendo tu ping ahora", () => api.quickPing());
    if (p) {
      setProbe(p);
      onDone();
    }
  };

  return (
    <section className="edge-light flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <span className="flex items-center gap-2 text-body-medium text-text-secondary">
          <StatusDot color={r?.inGame ? "green" : r?.running ? "yellow" : "gray"} />
          {status}
        </span>
        {r?.inGame ? <Chip variant="caption" color="neutral">{r.serverIp}</Chip> : <Chip variant="caption" color="lime">Realistic Street Soccer</Chip>}
      </div>
      <h2 className="mt-4 text-title-1-medium text-ink">{text.title}</h2>
      <p className="mt-1.5 text-body-regular text-text-secondary">{text.text}</p>
      <div className="mt-4 flex flex-col gap-2">
        {r?.inGame ? (
          <div className="grid grid-cols-2 gap-2 rounded-2lg bg-background-inner-default p-1.5 shadow-card">
            <Stat label="Ping ahora" value={fresh?.rttMs} tone="volt" />
            <Stat label="Tu mejor centro" value={best?.rttMs} tone="ink" />
          </div>
        ) : (
          <>
            <RegionSelect regions={regions} value={region} onChange={setRegion} isDisabled={!!task} />
            {region === PRIVATE ? (
              <TextField
                value={privateLink}
                onChange={setPrivateLink}
                isDisabled={!!task}
                className="flex flex-col rounded-2lg bg-background-inner-default px-4 py-2.5 shadow-card focus-within:ring-2 focus-within:ring-border-focus-ring"
              >
                <Label className="text-caption-1-medium text-ink-3">Enlace del servidor privado</Label>
                <Input placeholder="https://www.roblox.com/share?code=…" className="mt-0.5 w-full bg-transparent text-body-medium text-ink outline-none placeholder:text-ink-4" />
              </TextField>
            ) : null}
            {region === PRIVATE ? <RegionSelect variant="private" regions={regions} value={privateRegion} onChange={setPrivateRegion} isDisabled={!!task} /> : null}
          </>
        )}
      </div>
      <div className="mt-auto pt-4">
        {r?.inGame ? (
          <Button className="h-11 w-full text-[15px]" leadingIcon={IconSignal} disabled={!!task} onClick={measure}>
            {fresh ? "Medir otra vez" : "Medir mi ping ahora"}
          </Button>
        ) : (
          <Button
            className="h-11 w-full text-[15px]"
            leadingIcon={IconPlay}
            disabled={!!task || !r?.installed || r.running || needsLink}
            onClick={async () => {
              if (await start()) onDone();
            }}
          >
            Jugar
          </Button>
        )}
      </div>
    </section>
  );
}
