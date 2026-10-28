import { useState } from "react";
import { api, errorText } from "@/lib/api";
import { play, setSoundEnabled, soundEnabled } from "@/lib/sound";
import { installUpdate, updateLine } from "@/lib/updates";
import { useApp } from "@/state";
import { Card, PageHeader, StatRow } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Switch } from "@/components/ui/switch";
import { AppIcon } from "@/components/Logo";
import { IconDownload, IconImport, IconRefresh, IconRevert, IconSound } from "@/components/icons";

const licenses = [
  { name: "Intel PresentMon", license: "MIT", use: "Captura de frames por ETW" },
  { name: "BoardUI", license: "MIT", use: "Tarjetas, chips, botones e interruptores" },
  { name: "Recharts", license: "MIT", use: "Gráficas de ping y de centros de datos" },
  { name: "Gravity UI Icons", license: "MIT", use: "Iconografía" },
  { name: "React Aria Components", license: "Apache 2.0", use: "Accesibilidad de controles" },
  { name: "Motion", license: "MIT", use: "Animaciones" },
  { name: "Geist", license: "OFL 1.1", use: "Tipografía de interfaz" },
  { name: "Barlow Condensed", license: "OFL 1.1", use: "Tipografía de cifras" },
  { name: "Wails", license: "MIT", use: "Ventana nativa" },
];

const cues = [
  { cue: "tap" as const, label: "Pulsar" },
  { cue: "on" as const, label: "Activar" },
  { cue: "off" as const, label: "Quitar" },
  { cue: "success" as const, label: "Prueba lista" },
  { cue: "error" as const, label: "Error" },
];

export function Settings() {
  const { overview, run, task, toast, update, setUpdate } = useApp();
  const [sound, setSound] = useState(soundEnabled());
  const [checking, setChecking] = useState(false);
  const [exporting, setExporting] = useState(false);
  return (
    <div>
      <PageHeader eyebrow="Ajustes" title="Preferencias" />
      <div className="grid grid-cols-2 gap-4">
        <Card className="p-5">
          <div className="flex items-center gap-4">
            <AppIcon className="size-12" />
            <div className="min-w-0">
              <p className="text-[17px] font-semibold text-ink">Toque</p>
              <p className="text-caption-1-medium text-ink-3">Versión {overview?.version}</p>
            </div>
            <Chip variant="caption" color={overview?.elevated ? "lime" : "neutral"} className="ml-auto">
              {overview?.elevated ? "Administrador" : "Permisos limitados"}
            </Chip>
          </div>
          <div className="mt-4">
            <StatRow label="Roblox" value={overview?.roblox.installed ? overview.roblox.version.replace("version-", "") : "No instalado"} />
            <StatRow label="Datos y resultados" value="C:\ProgramData\Toque" />
            <StatRow label="Código y versiones" value="dimelim/toque" />
          </div>
        </Card>

        <Card className="flex flex-col p-5">
          <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
            <IconDownload className="size-4 text-ink-3" />
            Actualizaciones
          </p>
          <p className="mt-3 text-body-regular text-ink-2">{updateLine(update)}</p>
          {update?.notes && (update.state === "ready" || update.state === "downloading") ? <p className="mt-2 text-caption-1-medium text-ink-3">{update.notes}</p> : null}
          {update?.notes && update.state === "current" && update.latest === update.current ? <p className="mt-2 text-caption-1-medium text-ink-3">Novedades de esta versión. {update.notes}</p> : null}
          <div className="mt-3 flex flex-wrap gap-1.5">
            <Chip variant="caption" color="lime">
              Firma Ed25519
            </Chip>
            <Chip variant="caption" color="neutral">
              Revisa cada 30 minutos
            </Chip>
            <Chip variant="caption" color="neutral">
              Se instala al reabrir
            </Chip>
          </div>
          <div className="mt-auto pt-4">
            {update?.state === "ready" ? (
              <Button leadingIcon={IconDownload} disabled={!!task} onClick={() => installUpdate(toast)}>
                Instalar Toque {update.latest}
              </Button>
            ) : (
              <Button
                variant="secondary"
                leadingIcon={IconRefresh}
                disabled={checking || update?.state === "checking" || update?.state === "downloading"}
                onClick={async () => {
                  setChecking(true);
                  try {
                    setUpdate(await api.checkUpdate());
                  } catch (e) {
                    toast({ tone: "error", title: "No se pudo comprobar", detail: errorText(e) });
                  } finally {
                    setChecking(false);
                  }
                }}
              >
                Buscar actualizaciones
              </Button>
            )}
          </div>
        </Card>

        <Card className="flex flex-col p-5">
          <div className="flex items-start justify-between gap-6">
            <div>
              <p className="flex items-center gap-1.5 text-body-medium text-ink">
                <IconSound className="size-4 text-ink-3" />
                Sonidos
              </p>
              <p className="mt-1 text-body-regular text-ink-2">Tonos cortos al activar cambios y al terminar una prueba, generados al momento.</p>
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
          <div className="mt-auto flex flex-wrap gap-2 pt-4">
            {cues.map((c) => (
              <Button key={c.cue} size="small" variant="secondary" disabled={!sound} onClick={() => play(c.cue)}>
                {c.label}
              </Button>
            ))}
          </div>
        </Card>

        <Card className="flex flex-col p-5">
          <p className="flex items-center gap-1.5 text-body-medium text-ink">
            <IconRevert className="size-4 text-ink-3" />
            Restaurar
          </p>
          <p className="mt-1 text-body-regular text-ink-2">Deshace todos los cambios que Toque hizo en Windows y en Roblox y deja cada ajuste como estaba antes de instalarlo. Desinstalar Toque hace lo mismo.</p>
          <div className="mt-auto pt-4">
            <Button
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
          </div>
        </Card>

        <Card className="col-span-2 flex items-center justify-between gap-6 p-5">
          <div>
            <p className="flex items-center gap-1.5 text-body-medium text-ink">
              <IconImport className="size-4 text-ink-3" />
              Informe de errores
            </p>
            <p className="mt-1 max-w-[640px] text-body-regular text-ink-2">Guarda en tu escritorio un .zip con los errores de Toque, tus flags y los últimos registros de Roblox. Envíalo a quien te da soporte. Incluye tu usuario de Roblox y las IP de los servidores en los que jugaste.</p>
          </div>
          <div className="shrink-0">
            <Button
              variant="secondary"
              leadingIcon={IconImport}
              disabled={exporting}
              onClick={async () => {
                setExporting(true);
                try {
                  const path = await api.exportReport();
                  toast({ tone: "ok", title: "Informe guardado en el escritorio", detail: path.split("\\").pop() });
                } catch (e) {
                  toast({ tone: "error", title: "No se pudo crear el informe", detail: errorText(e) });
                } finally {
                  setExporting(false);
                }
              }}
            >
              Crear informe
            </Button>
          </div>
        </Card>
      </div>

      <Card className="mt-4 p-5">
        <p className="text-body-medium text-text-secondary">Software de terceros</p>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {licenses.map((l) => (
            <div key={l.name} className="flex items-center justify-between gap-3 rounded-xl bg-background-inner-default px-3 py-2.5 shadow-card">
              <div className="min-w-0">
                <p className="truncate text-body-medium text-ink">{l.name}</p>
                <p className="truncate text-caption-1-medium text-ink-3">{l.use}</p>
              </div>
              <Chip variant="caption" color="neutral">
                {l.license}
              </Chip>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
