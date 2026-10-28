import { useState } from "react";
import { Input, Label, TextField } from "react-aria-components";
import { api } from "@/lib/api";
import { num } from "@/lib/format";
import { t } from "@/lib/i18n";
import { PRIVATE, bestRegion, regionCity } from "@/lib/regions";
import { play } from "@/lib/sound";
import type { Region, ServerProbe, WatchInfo } from "@/lib/types";
import { usePlay } from "@/hooks/use-play";
import { useApp } from "@/state";
import { cx } from "@/utils/cx";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { StatusDot } from "@/components/ui/status-dot";
import { RegionSelect } from "@/components/RegionSelect";
import { IconPlay, IconSearch, IconSignal } from "@/components/icons";

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

function intro(inGame: boolean, region: string, privateRegion: string, watching: boolean) {
  const city = region && region !== PRIVATE ? regionCity(region) : "";
  if (watching) return { title: t("Vigilando {city}", { city }), text: t("Toque revisa la lista cada 25 segundos y te mete en cuanto haya una partida de {city} con sitio, aunque estés jugando en otra.", { city }) };
  if (inGame) return { title: t("Ya estás jugando"), text: t("Mide el ping de este servidor o cámbiate a otro con la región que elijas.") };
  if (region === PRIVATE && privateRegion) {
    const target = regionCity(privateRegion);
    return { title: t("Tu servidor privado en {city}", { city: target }), text: t("Roblox elige la ciudad al abrirlo, cerca de quien entra primero. Toque entra, comprueba dónde quedó y, si no es {city}, te dice qué hacer.", { city: target }) };
  }
  if (region === PRIVATE) return { title: t("Jugar en tu servidor privado"), text: t("Pega el enlace de invitación. Toque lo guarda solo en este PC y entra directo.") };
  if (city) return { title: t("Jugar en {city}", { city }), text: t("Toque busca con RoValra una partida de {city} con sitio y entra directo. Si no hay ninguna, déjalo vigilando.", { city }) };
  return { title: t("Jugar en un servidor cercano"), text: t("Roblox te manda a un servidor lejano en una de cada seis partidas. Toque lo comprueba al entrar y, si pasa, te mueve a uno cercano.") };
}

function minutesSince(iso: string) {
  return Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000));
}

function WatchStatus({ watch }: { watch: WatchInfo }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-2lg bg-background-inner-default px-4 py-3 shadow-card">
      <span className="flex items-center gap-2 text-body-medium text-ink">
        <StatusDot color="green" />
        {t("{n} partidas revisadas", { n: num(watch.checked) })}
      </span>
      <span className="text-caption-1-medium text-ink-3">{t("hace {n} min", { n: minutesSince(watch.since) })}</span>
    </div>
  );
}

export function PlayCard({ regions, onDone }: { regions: Region[]; onDone: () => void }) {
  const { overview, run, task, region, setRegion, privateLink, setPrivateLink, privateRegion, setPrivateRegion, watch } = useApp();
  const start = usePlay();
  const [probe, setProbe] = useState<ServerProbe | null>(null);
  const r = overview?.roblox;
  const inGame = !!r?.inGame;
  const running = !!r?.running;
  const best = bestRegion(regions);
  const fresh = probe && inGame && probe.jobId === r?.jobId ? probe : null;
  const status = !r?.installed ? t("Roblox no está instalado") : inGame ? t("En partida") : running ? t("Roblox abierto") : t("Roblox cerrado");
  const watching = !!watch?.active && watch.region === region;
  const text = intro(inGame, region, privateRegion, watching);
  const needsLink = region === PRIVATE && !privateLink.trim();
  const canWatch = !!region && region !== PRIVATE;

  const measure = async () => {
    const p = await run("ping", t("Midiendo tu ping ahora"), () => api.quickPing());
    if (p) {
      setProbe(p);
      onDone();
    }
  };

  return (
    <section className="flex min-w-0 flex-col rounded-2xl bg-background-secondary-default p-5 shadow-card">
      <div className="flex items-center justify-between gap-3">
        <span className="flex items-center gap-2 text-body-medium text-text-secondary">
          <StatusDot color={inGame ? "green" : running ? "yellow" : "gray"} />
          {status}
        </span>
        {inGame ? <Chip variant="caption" color="neutral">{r?.serverIp}</Chip> : <Chip variant="caption" color="lime">Realistic Street Soccer</Chip>}
      </div>
      <h2 className="mt-4 text-title-1-medium text-ink">{text.title}</h2>
      <p className="mt-1.5 text-body-regular text-text-secondary">{text.text}</p>
      <div className="mt-4 flex flex-col gap-2">
        {inGame ? (
          <div className="flex items-center gap-1 rounded-2lg bg-background-inner-default p-1.5 shadow-card">
            <Stat label={t("Ping ahora")} value={fresh?.rttMs} tone="volt" />
            <Stat label={t("Tu mejor centro")} value={best?.rttMs} tone="ink" />
            <Button size="small" variant="secondary" className="mr-1.5 ml-auto" leadingIcon={IconSignal} disabled={!!task} onClick={measure}>
              {fresh ? t("Otra vez") : t("Medir")}
            </Button>
          </div>
        ) : null}
        <RegionSelect regions={regions} value={region} onChange={setRegion} isDisabled={!!task || watching} />
        {region === PRIVATE ? (
          <>
            <TextField
              value={privateLink}
              onChange={setPrivateLink}
              isDisabled={!!task}
              className="flex flex-col rounded-2lg bg-background-inner-default px-4 py-2.5 shadow-card focus-within:ring-2 focus-within:ring-border-focus-ring"
            >
              <Label className="text-caption-1-medium text-ink-3">{t("Enlace del servidor privado")}</Label>
              <Input placeholder="https://www.roblox.com/share?code=…" className="mt-0.5 w-full bg-transparent text-body-medium text-ink outline-none placeholder:text-ink-4" />
            </TextField>
            <RegionSelect variant="private" regions={regions} value={privateRegion} onChange={setPrivateRegion} isDisabled={!!task} />
          </>
        ) : null}
        {watching && watch ? <WatchStatus watch={watch} /> : null}
      </div>
      <div className="mt-auto pt-4">
        {watching ? (
          <Button
            variant="secondary"
            className="h-11 w-full text-[15px]"
            onClick={() => {
              play("off");
              api.stopWatch();
            }}
          >
            {t("Dejar de vigilar")}
          </Button>
        ) : (
          <div className={cx("grid gap-2", canWatch && "grid-cols-[1fr_auto]")}>
            <Button
              className="h-11 w-full text-[15px]"
              leadingIcon={IconPlay}
              disabled={!!task || !r?.installed || needsLink}
              onClick={async () => {
                if (await start()) onDone();
              }}
            >
              {running ? t("Cambiar de servidor") : t("Jugar")}
            </Button>
            {canWatch ? (
              <Button
                variant="secondary"
                className="h-11 text-[15px]"
                leadingIcon={IconSearch}
                disabled={!r?.installed}
                onClick={() => {
                  play("on");
                  api.watchRegion(region);
                }}
              >
                {t("Vigilar")}
              </Button>
            ) : null}
          </div>
        )}
        {running && !watching ? (
          <p className="mt-2 text-caption-1-medium text-ink-3">{inGame ? t("Al cambiar de servidor saldrás de la partida actual.") : t("Roblox está abierto. Toque lo cerrará para entrar en el servidor elegido.")}</p>
        ) : null}
      </div>
    </section>
  );
}
