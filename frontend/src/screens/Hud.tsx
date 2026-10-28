import { useCallback, useEffect, useState } from "react";
import { Label, Slider, SliderOutput, SliderThumb, SliderTrack } from "react-aria-components";
import { cx } from "@/utils/cx";
import { api, errorText } from "@/lib/api";
import { play } from "@/lib/sound";
import type { HudConfig, HudLayout } from "@/lib/types";
import { useApp } from "@/state";
import { Card, Eyebrow, PageHeader } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { HudEditor } from "./hud/Editor";

function Row({ title, text, on, disabled, onChange }: { title: string; text: string; on: boolean; disabled?: boolean; onChange: (v: boolean) => void }) {
  return (
    <div className="flex items-center justify-between gap-4 py-3 shadow-[inset_0_-1px_0_var(--color-line)] last:shadow-none">
      <div className="min-w-0">
        <p className="text-body-medium text-ink">{title}</p>
        <p className="text-caption-1-medium text-ink-3">{text}</p>
      </div>
      <Switch size="md" isSelected={on} isDisabled={disabled} onChange={onChange} aria-label={title} />
    </div>
  );
}

function Range({
  label,
  value,
  min,
  max,
  step,
  format,
  disabled,
  onCommit,
}: {
  label: string;
  value: number;
  min: number;
  max: number;
  step: number;
  format: (v: number) => string;
  disabled?: boolean;
  onCommit: (v: number) => void;
}) {
  const [local, setLocal] = useState(value);
  useEffect(() => setLocal(value), [value]);
  return (
    <Slider
      className="flex flex-col gap-2"
      minValue={min}
      maxValue={max}
      step={step}
      value={local}
      onChange={(v) => setLocal(v as number)}
      onChangeEnd={(v) => onCommit(v as number)}
      isDisabled={disabled}
    >
      <div className="flex items-baseline justify-between">
        <Label className="text-body-medium text-ink">{label}</Label>
        <SliderOutput className="num text-body-medium text-volt">{format(local)}</SliderOutput>
      </div>
      <SliderTrack className="relative h-6 w-full cursor-pointer data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50">
        {({ state }) => (
          <>
            <div className="absolute top-1/2 h-1 w-full -translate-y-1/2 rounded-full bg-surface-3" />
            <div className="absolute top-1/2 h-1 -translate-y-1/2 rounded-full bg-volt" style={{ width: `${state.getThumbPercent(0) * 100}%` }} />
            <SliderThumb className="top-1/2 size-[18px] rounded-full bg-ink shadow-[0_1px_4px_rgb(0_0_0/0.45)] outline-none transition-[scale] duration-150 data-[dragging]:scale-110 data-[focus-visible]:ring-2 data-[focus-visible]:ring-border-focus-ring" />
          </>
        )}
      </SliderTrack>
    </Slider>
  );
}

const corners = [
  { id: "tl", label: "Arriba izq." },
  { id: "tr", label: "Arriba der." },
  { id: "bl", label: "Abajo izq." },
  { id: "br", label: "Abajo der." },
];

export function Hud() {
  const { overview, toast } = useApp();
  const [cfg, setCfg] = useState<HudConfig | null>(null);
  const [layout, setLayout] = useState<HudLayout | null>(null);
  const load = useCallback(async () => {
    try {
      setCfg(await api.hudConfig());
    } catch {
      return;
    }
  }, []);
  useEffect(() => {
    load();
    api.previewHud(true).catch(() => undefined);
    return () => {
      api.previewHud(false).catch(() => undefined);
    };
  }, [load]);
  useEffect(() => {
    if (!cfg) return;
    api.hudLayout().then(setLayout).catch(() => undefined);
  }, [cfg?.scale, cfg?.showKeys, cfg?.showMouse, cfg?.showPing, !!cfg]);

  const commit = async (next: HudConfig) => {
    setCfg(next);
    try {
      setCfg(await api.setHudConfig(next));
      api.hudLayout().then(setLayout).catch(() => undefined);
    } catch (e) {
      toast({ tone: "error", title: "No pude cambiar el HUD", detail: errorText(e) });
      load();
    }
  };
  const patch = (p: Partial<HudConfig>) => {
    if (!cfg) return;
    commit({ ...cfg, ...p });
  };
  const place = async (x: number, y: number, save: boolean) => {
    try {
      const next = await api.placeHud(x, y, save);
      if (save) setCfg(next);
    } catch (e) {
      if (save) toast({ tone: "error", title: "No pude mover el HUD", detail: errorText(e) });
    }
  };
  const corner = async (id: string) => {
    try {
      play("tap");
      setCfg(await api.moveHud(id));
    } catch (e) {
      toast({ tone: "error", title: "No pude mover el HUD", detail: errorText(e) });
    }
  };

  const on = !!cfg?.enabled;
  const running = !!overview?.roblox.running;

  return (
    <div>
      <PageHeader eyebrow="HUD" title="Tu HUD en el juego">
        <Switch
          size="lg"
          isSelected={on}
          isDisabled={!cfg}
          onChange={(v) => {
            play(v ? "on" : "off");
            patch({ enabled: v });
          }}
          aria-label="Mostrar el HUD"
        >
          {on ? "Encendido" : "Apagado"}
        </Switch>
      </PageHeader>

      <Card className="mb-8 p-5">
        <Eyebrow>Qué es</Eyebrow>
        <p className="mt-2 max-w-[80ch] text-body-regular text-ink-2">
          Una capa encima de Roblox con tu ping, las teclas W A S D y los botones del ratón, que se encienden al pulsarlas. Útil para grabar o para ver tus entradas mientras juegas. Solo aparece mientras Roblox está en primer plano y solo se redibuja cuando algo cambia, para no quitarte FPS.
        </p>
        <p className="mt-2 text-caption-1-medium text-ink-3">
          {running ? "Roblox está abierto. Juega en ventana o pantalla completa sin bordes para verlo encima." : "Se ve encima de Roblox en ventana o pantalla completa sin bordes, no en pantalla completa exclusiva."}
        </p>
      </Card>

      <div className={cx("grid grid-cols-2 gap-3", !on && "pointer-events-none opacity-50")}>
        <Card className="p-5">
          <h2 className="mb-1 text-[15px] font-semibold text-ink">Qué mostrar</h2>
          <Row title="Ping" text="Tu ping al servidor" on={!!cfg?.showPing} disabled={!on} onChange={(v) => patch({ showPing: v })} />
          <Row title="Teclas W A S D" text="Con espacio y shift, animadas" on={!!cfg?.showKeys} disabled={!on} onChange={(v) => patch({ showKeys: v })} />
          <Row title="Ratón" text="Clics y movimiento" on={!!cfg?.showMouse} disabled={!on} onChange={(v) => patch({ showMouse: v })} />
        </Card>

        <Card className="flex flex-col gap-5 p-5">
          <h2 className="text-[15px] font-semibold text-ink">Tamaño</h2>
          <Range label="Escala" value={cfg?.scale ?? 1} min={0.6} max={2.2} step={0.1} format={(v) => `${Math.round(v * 100)} %`} disabled={!on} onCommit={(v) => patch({ scale: v })} />
          <Range label="Opacidad" value={cfg?.opacity ?? 90} min={30} max={100} step={5} format={(v) => `${Math.round(v)} %`} disabled={!on} onCommit={(v) => patch({ opacity: Math.round(v) })} />
        </Card>

        <Card className="col-span-2 p-5">
          <h2 className="mb-1 text-[15px] font-semibold text-ink">Posición</h2>
          <p className="mb-4 text-caption-1-medium text-ink-3">Arrastra el HUD en la miniatura de tu pantalla. Se pega a los bordes y al centro, y mientras estás aquí lo ves también encima del escritorio.</p>
          <div className="grid grid-cols-[minmax(0,1fr)_200px] items-start gap-6">
            {cfg && layout ? <HudEditor cfg={cfg} layout={layout} disabled={!on} onPlace={place} /> : <div />}
            <div>
              <p className="mb-2 text-caption-1-medium text-ink-3">Esquinas</p>
              <div className="grid grid-cols-2 gap-2">
                {corners.map((c) => (
                  <Button key={c.id} size="small" variant="secondary" disabled={!on} onClick={() => corner(c.id)}>
                    {c.label}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
