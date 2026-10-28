import { useCallback, useEffect, useState } from "react";
import { api, errorText } from "@/lib/api";
import { t } from "@/lib/i18n";
import { play, setSoundEnabled, soundEnabled } from "@/lib/sound";
import { installUpdate, updateLine } from "@/lib/updates";
import { useApp } from "@/state";
import { Card, PageHeader, StatRow } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Switch } from "@/components/ui/switch";
import { AppIcon } from "@/components/Logo";
import { IconDownload, IconImport, IconRefresh, IconRevert, IconSound } from "@/components/icons";
import { LanguageCard } from "./settings/LanguageCard";

const licenses = () => [
  { name: "Intel PresentMon", license: "MIT", use: t("Captura de frames por ETW") },
  { name: "BoardUI", license: "MIT", use: t("Tarjetas, chips, botones e interruptores") },
  { name: "Recharts", license: "MIT", use: t("Gráficas de ping y de centros de datos") },
  { name: "Gravity UI Icons", license: "MIT", use: t("Iconografía") },
  { name: "React Aria Components", license: "Apache 2.0", use: t("Accesibilidad de controles") },
  { name: "Motion", license: "MIT", use: t("Animaciones") },
  { name: "Geist", license: "OFL 1.1", use: t("Tipografía de interfaz") },
  { name: "Barlow Condensed", license: "OFL 1.1", use: t("Tipografía de cifras") },
  { name: "Wails", license: "MIT", use: t("Ventana nativa") },
];

const cues = () => [
  { cue: "tap" as const, label: t("Pulsar") },
  { cue: "on" as const, label: t("Activar") },
  { cue: "off" as const, label: t("Quitar") },
  { cue: "success" as const, label: t("Prueba lista") },
  { cue: "error" as const, label: t("Error") },
];

export function Settings() {
  const { overview, run, task, toast, update, setUpdate } = useApp();
  const [sound, setSound] = useState(soundEnabled());
  const [checking, setChecking] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [changes, setChanges] = useState<{ id: string; title: string }[]>([]);
  const loadChanges = useCallback(() => {
    api
      .activeChanges()
      .then(setChanges)
      .catch(() => undefined);
  }, []);
  useEffect(() => {
    loadChanges();
  }, [loadChanges]);
  return (
    <div>
      <PageHeader eyebrow={t("Ajustes")} title={t("Preferencias")} />
      <div className="grid grid-cols-2 gap-4">
        <Card className="p-5">
          <div className="flex items-center gap-4">
            <AppIcon className="size-12" />
            <div className="min-w-0">
              <p className="text-[17px] font-semibold text-ink">Toque</p>
              <p className="text-caption-1-medium text-ink-3">{t("Versión {version}", { version: overview?.version ?? "" })}</p>
            </div>
            <Chip variant="caption" color={overview?.elevated ? "lime" : "neutral"} className="ml-auto">
              {overview?.elevated ? t("Administrador") : t("Permisos limitados")}
            </Chip>
          </div>
          <div className="mt-4">
            <StatRow label="Roblox" value={overview?.roblox.installed ? overview.roblox.version.replace("version-", "") : t("No instalado")} />
            <StatRow label={t("Datos y resultados")} value="C:\ProgramData\Toque" />
            <StatRow label={t("Código y versiones")} value="dimelim/toque" />
          </div>
        </Card>

        <Card className="flex flex-col p-5">
          <p className="flex items-center gap-1.5 text-body-medium text-text-secondary">
            <IconDownload className="size-4 text-ink-3" />
            {t("Actualizaciones")}
          </p>
          <p className="mt-3 text-body-regular text-ink-2">{updateLine(update)}</p>
          {update?.notes && (update.state === "ready" || update.state === "downloading") ? <p className="mt-2 text-caption-1-medium text-ink-3">{update.notes}</p> : null}
          {update?.notes && update.state === "current" && update.latest === update.current ? <p className="mt-2 text-caption-1-medium text-ink-3">{t("Novedades de esta versión. {notes}", { notes: update.notes })}</p> : null}
          <div className="mt-3 flex flex-wrap gap-1.5">
            <Chip variant="caption" color="lime">
              {t("Firma Ed25519")}
            </Chip>
            <Chip variant="caption" color="neutral">
              {t("Revisa cada 30 minutos")}
            </Chip>
            <Chip variant="caption" color="neutral">
              {t("Se instala al reabrir")}
            </Chip>
          </div>
          <div className="mt-auto pt-4">
            {update?.state === "ready" ? (
              <Button leadingIcon={IconDownload} disabled={!!task} onClick={() => installUpdate(toast)}>
                {t("Instalar Toque {version}", { version: update.latest })}
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
                    toast({ tone: "error", title: t("No se pudo comprobar"), detail: errorText(e) });
                  } finally {
                    setChecking(false);
                  }
                }}
              >
                {t("Buscar actualizaciones")}
              </Button>
            )}
          </div>
        </Card>

        <Card className="flex flex-col p-5">
          <div className="flex items-start justify-between gap-6">
            <div>
              <p className="flex items-center gap-1.5 text-body-medium text-ink">
                <IconSound className="size-4 text-ink-3" />
                {t("Sonidos")}
              </p>
              <p className="mt-1 text-body-regular text-ink-2">{t("Tonos cortos al activar cambios y al terminar una prueba, generados al momento.")}</p>
            </div>
            <Switch
              size="md"
              isSelected={sound}
              onChange={(v) => {
                setSound(v);
                setSoundEnabled(v);
                if (v) play("on");
              }}
              aria-label={t("Sonidos")}
            />
          </div>
          <div className="mt-auto flex flex-wrap gap-2 pt-4">
            {cues().map((c) => (
              <Button key={c.cue} size="small" variant="secondary" disabled={!sound} onClick={() => play(c.cue)}>
                {c.label}
              </Button>
            ))}
          </div>
        </Card>

        <Card className="flex flex-col p-5">
          <p className="flex items-center gap-1.5 text-body-medium text-ink">
            <IconRevert className="size-4 text-ink-3" />
            {t("Restaurar")}
          </p>
          <p className="mt-1 text-body-regular text-ink-2">{t("Deshace todos los cambios que Toque hizo en Windows y en Roblox y deja cada ajuste como estaba antes de instalarlo. Desinstalar Toque hace lo mismo.")}</p>
          <div className="mt-3 rounded-lg bg-surface-1 px-3 py-2 shadow-[inset_0_0_0_1px_var(--color-line)]">
            <p className="text-caption-1-medium text-ink-3">{changes.length === 0 ? t("Ahora mismo Toque no tiene ningún cambio activo.") : changes.length === 1 ? t("1 cambio activo ahora mismo") : t("{n} cambios activos ahora mismo", { n: changes.length })}</p>
            {changes.length > 0 ? (
              <ul className="mt-1.5 flex max-h-[132px] flex-col gap-1 overflow-y-auto">
                {changes.map((c) => (
                  <li key={c.id} className="flex items-center gap-2 text-caption-1-medium text-ink-2">
                    <span className="size-1.5 shrink-0 rounded-full bg-volt" />
                    <span className="truncate">{c.title}</span>
                  </li>
                ))}
              </ul>
            ) : null}
          </div>
          <div className="mt-auto pt-4">
            <Button
              variant="secondary"
              leadingIcon={IconRevert}
              disabled={!!task}
              onClick={async () => {
                await run("revert", t("Revirtiendo todos los cambios"), () => api.revertAll());
                toast({ tone: "ok", title: t("Todo vuelve a como estaba") });
                loadChanges();
              }}
            >
              {t("Revertir todos los cambios")}
            </Button>
          </div>
        </Card>

        <LanguageCard />

        <Card className="flex flex-col p-5">
          <div>
            <p className="flex items-center gap-1.5 text-body-medium text-ink">
              <IconImport className="size-4 text-ink-3" />
              {t("Informe de errores")}
            </p>
            <p className="mt-1 text-body-regular text-ink-2">{t("Guarda en tu escritorio un .zip con los errores de Toque, tus flags y los últimos registros de Roblox. Envíalo a quien te da soporte. Incluye tu usuario de Roblox y las IP de los servidores en los que jugaste.")}</p>
          </div>
          <div className="mt-auto pt-4">
            <Button
              variant="secondary"
              leadingIcon={IconImport}
              disabled={exporting}
              onClick={async () => {
                setExporting(true);
                try {
                  const path = await api.exportReport();
                  toast({ tone: "ok", title: t("Informe guardado en el escritorio"), detail: path.split("\\").pop() });
                } catch (e) {
                  toast({ tone: "error", title: t("No se pudo crear el informe"), detail: errorText(e) });
                } finally {
                  setExporting(false);
                }
              }}
            >
              {t("Crear informe")}
            </Button>
          </div>
        </Card>
      </div>

      <Card className="mt-4 p-5">
        <p className="text-body-medium text-text-secondary">{t("Software de terceros")}</p>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {licenses().map((l) => (
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
