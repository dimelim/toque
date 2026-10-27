import { api, errorText } from "./api";
import { when } from "./format";
import type { UpdateInfo } from "./types";

type Notify = (t: { tone: "ok" | "error" | "info"; title: string; detail?: string }) => void;

export async function installUpdate(notify: Notify) {
  try {
    await api.applyUpdate();
  } catch (e) {
    notify({ tone: "error", title: "No se pudo actualizar", detail: errorText(e) });
  }
}

export function updateLine(u: UpdateInfo | null) {
  switch (u?.state) {
    case "checking":
      return "Buscando una versión nueva…";
    case "downloading":
      return `Descargando Toque ${u.latest} y comprobando su firma…`;
    case "ready":
      return `Toque ${u.latest} está descargada y verificada. Se instala sola la próxima vez que abras Toque, o ahora mismo si prefieres.`;
    case "current":
      return `Tienes la última versión. Última comprobación ${when(u.checked)}`;
    case "error":
      return `No se pudo comprobar si hay versión nueva (${u.error}).`;
    default:
      return "Toque busca versiones nuevas al abrirse y cada 30 minutos, y solo instala las que llevan su firma.";
  }
}
