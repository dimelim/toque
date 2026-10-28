import { api, errorText } from "./api";
import { when } from "./format";
import { t } from "./i18n";
import type { UpdateInfo } from "./types";

type Notify = (t: { tone: "ok" | "error" | "info"; title: string; detail?: string }) => void;

export async function installUpdate(notify: Notify) {
  try {
    await api.applyUpdate();
  } catch (e) {
    notify({ tone: "error", title: t("No se pudo actualizar"), detail: errorText(e) });
  }
}

export function updateLine(u: UpdateInfo | null) {
  switch (u?.state) {
    case "checking":
      return t("Buscando una versión nueva…");
    case "downloading":
      return t("Descargando Toque {version} y comprobando su firma…", { version: u.latest });
    case "ready":
      return t("Toque {version} está descargada y verificada. Se instala sola la próxima vez que abras Toque, o ahora mismo si prefieres.", { version: u.latest });
    case "current":
      return t("Tienes la última versión. Última comprobación {when}", { when: when(u.checked) });
    case "error":
      return t("No se pudo comprobar si hay versión nueva ({error}).", { error: u.error });
    default:
      return t("Toque busca versiones nuevas al abrirse y cada 30 minutos, y solo instala las que llevan su firma.");
  }
}
