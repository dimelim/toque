import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { play, setSoundEnabled, soundEnabled } from "@/lib/sound";
import { useApp } from "@/state";
import { Card, Eyebrow, PageHeader, StatRow } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { AppIcon } from "@/components/Logo";
import { IconRevert } from "@/components/icons";
import * as Service from "@wails/go/app/Service";

const licenses = [
  { name: "Intel PresentMon", license: "MIT", use: "Captura de frames por ETW" },
  { name: "BoardUI", license: "MIT", use: "Botones, interruptores y tooltips" },
  { name: "Gravity UI Icons", license: "MIT", use: "Iconografía" },
  { name: "React Aria Components", license: "Apache 2.0", use: "Accesibilidad de controles" },
  { name: "Motion", license: "MIT", use: "Animaciones" },
  { name: "Geist", license: "OFL 1.1", use: "Tipografía de interfaz" },
  { name: "Barlow Condensed", license: "OFL 1.1", use: "Tipografía de cifras" },
  { name: "Wails", license: "MIT", use: "Ventana nativa" },
];

export function Settings() {
  const { overview, run, task, toast } = useApp();
  const [sound, setSound] = useState(soundEnabled());
  const [agent, setAgent] = useState(false);
  useEffect(() => {
    Service.AgentEnabled().then(setAgent).catch(() => setAgent(false));
  }, []);
  return (
    <div>
      <PageHeader eyebrow="Ajustes" title="Preferencias" />
      <div className="grid grid-cols-[1.2fr_1fr] gap-4">
        <div className="flex flex-col gap-4">
          <Card className="p-5">
            <Eyebrow>Durante la partida</Eyebrow>
            <div className="mt-4 flex items-start justify-between gap-6">
              <div>
                <p className="text-body-medium text-ink">Activar Toque al iniciar Windows</p>
                <p className="mt-1 max-w-[52ch] text-body-regular text-ink-2">
                  Un agente ligero, sin ventana y con menos de 15 MB de memoria, aplica las optimizaciones en vivo solo mientras Roblox está abierto y las retira al cerrarlo.
                </p>
              </div>
              <Switch
                size="md"
                isSelected={agent}
                onChange={async (v) => {
                  try {
                    await Service.SetAgentEnabled(v);
                    setAgent(v);
                    play(v ? "on" : "off");
                  } catch {
                    play("error");
                  }
                }}
                aria-label="Activar Toque al iniciar Windows"
              />
            </div>
          </Card>
          <Card className="p-5">
            <Eyebrow>Interfaz</Eyebrow>
            <div className="mt-4 flex items-start justify-between gap-6">
              <div>
                <p className="text-body-medium text-ink">Sonidos</p>
                <p className="mt-1 text-body-regular text-ink-2">Pequeños tonos al activar cambios y al terminar una prueba.</p>
              </div>
              <Switch
                size="md"
                isSelected={sound}
                onChange={(v) => {
                  setSound(v);
                  setSoundEnabled(v);
                  if (v) play("on");
                }}
                aria-label="Sonidos"
              />
            </div>
          </Card>
          <Card className="p-5">
            <Eyebrow>Restaurar</Eyebrow>
            <p className="mt-3 max-w-[60ch] text-body-regular text-ink-2">
              Deshace todos los cambios que Toque hizo en Windows y en Roblox y deja cada ajuste como estaba antes de instalarlo. Desinstalar Toque hace lo mismo.
            </p>
            <Button
              className="mt-4"
              variant="secondary"
              leadingIcon={IconRevert}
              disabled={!!task}
              onClick={async () => {
                await run("revert", "Revirtiendo todos los cambios", () => api.revertAll());
                toast({ tone: "ok", title: "Todo vuelve a como estaba" });
              }}
            >
              Revertir todos los cambios
            </Button>
          </Card>
        </div>
        <div className="flex flex-col gap-4">
          <Card className="p-5">
            <div className="flex items-center gap-4">
              <AppIcon className="size-12" />
              <div>
                <p className="text-[17px] font-semibold text-ink">Toque</p>
                <p className="text-caption-1-medium text-ink-3">Versión {overview?.version}</p>
              </div>
            </div>
            <div className="mt-4">
              <StatRow label="Permisos" value={overview?.elevated ? "Administrador" : "Limitados"} />
              <StatRow label="Roblox" value={overview?.roblox.installed ? overview.roblox.version.replace("version-", "") : "No instalado"} />
              <StatRow label="Datos y resultados" value="C:\ProgramData\Toque" />
            </div>
          </Card>
          <Card className="p-5">
            <Eyebrow>Software de terceros</Eyebrow>
            <div className="mt-3">
              {licenses.map((l) => (
                <StatRow key={l.name} label={l.name} value={l.license} hint={l.use} />
              ))}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
