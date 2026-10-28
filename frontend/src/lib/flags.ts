import { share } from "./format";
import { t } from "./i18n";

export type FlagValues = Record<string, string>;

export interface FlagOption {
  id: string;
  label: string;
  values: FlagValues;
}

interface GroupBase {
  id: string;
  title: string;
  text: string;
  keys: string[];
  note?: string;
}

export type FlagGroup =
  | (GroupBase & { control: "switch"; on: FlagValues })
  | (GroupBase & { control: "choice"; options: FlagOption[] })
  | (GroupBase & { control: "level"; max: number });

export interface FlagSection {
  id: string;
  title: string;
  hint: string;
  groups: FlagGroup[];
}

const D3D11 = "FFlagDebugGraphicsPreferD3D11";
const VULKAN = "FFlagDebugGraphicsPreferVulkan";
const OPENGL = "FFlagDebugGraphicsPreferOpenGL";
const TEXTURES_ON = "DFFlagTextureQualityOverrideEnabled";
const TEXTURES = "DFIntTextureQualityOverride";
const MSAA = "FIntDebugForceMSAASamples";
const QUALITY = "DFIntDebugFRMQualityLevelOverride";
const VOXEL = "DFFlagDebugPauseVoxelizer";
const SKY = "FFlagDebugSkyGray";
const GRASS_MIN = "FIntFRMMinGrassDistance";
const GRASS_MAX = "FIntFRMMaxGrassDistance";
const GRASS_MOTION = "FIntGrassMovementReducedMotionFactor";
const ALT_ENTER = "FFlagHandleAltEnterFullscreenManually";
const DPI = "DFFlagDisableDPIScale";
const CSG = [
  "DFIntCSGLevelOfDetailSwitchingDistance",
  "DFIntCSGLevelOfDetailSwitchingDistanceL12",
  "DFIntCSGLevelOfDetailSwitchingDistanceL23",
  "DFIntCSGLevelOfDetailSwitchingDistanceL34",
];

const automatic = (label = t("Automática")): FlagOption => ({ id: "auto", label, values: {} });
const every = (keys: string[], value: string) => Object.fromEntries(keys.map((k) => [k, value]));

export const flagSections = (): FlagSection[] => [
  {
    id: "render",
    title: t("Render"),
    hint: t("Cómo dibuja Roblox cada frame"),
    groups: [
      {
        id: "api",
        control: "choice",
        title: t("API gráfica"),
        text: t("El motor con el que Roblox dibuja. Direct3D 11 entrega cada frame directo a la pantalla, y Vulkan pasa antes por el compositor de Windows."),
        keys: [D3D11, VULKAN, OPENGL],
        options: [
          automatic(),
          { id: "d3d11", label: "Direct3D 11", values: { [D3D11]: "True" } },
          { id: "vulkan", label: "Vulkan", values: { [VULKAN]: "True" } },
          { id: "opengl", label: "OpenGL", values: { [OPENGL]: "True" } },
        ],
        note: t("En las pruebas de Toque, Vulkan más que duplicó la latencia de render frente a Direct3D 11."),
      },
      {
        id: "quality",
        control: "level",
        title: t("Calidad gráfica forzada"),
        text: t("Fija la calidad de los gráficos entre 1 y 21, pase lo que pase en el menú de Roblox. No cambia la distancia de dibujado."),
        keys: [QUALITY],
        max: 21,
      },
      {
        id: "textures",
        control: "choice",
        title: t("Calidad de texturas"),
        text: t("Fija la resolución de las texturas. Con menos resolución ocupan menos memoria de vídeo y se ven más borrosas de cerca."),
        keys: [TEXTURES_ON, TEXTURES],
        options: [
          automatic(),
          ...[t("Mínima"), t("Baja"), t("Media"), t("Máxima")].map((label, level) => ({
            id: `level${level}`,
            label,
            values: { [TEXTURES_ON]: "True", [TEXTURES]: String(level) },
          })),
        ],
        note: t("En las pruebas de Toque, bajar las texturas no redujo la RAM de Roblox."),
      },
      {
        id: "msaa",
        control: "choice",
        title: t("Antialiasing (MSAA)"),
        text: t("Suaviza los bordes dibujando varias muestras por píxel. Cada muestra extra cuesta tiempo de GPU."),
        keys: [MSAA],
        options: [
          automatic(t("Automático")),
          { id: "1", label: t("Sin suavizado"), values: { [MSAA]: "1" } },
          { id: "2", label: "2×", values: { [MSAA]: "2" } },
          { id: "4", label: "4×", values: { [MSAA]: "4" } },
        ],
      },
      {
        id: "voxel",
        control: "switch",
        title: t("Iluminación estática"),
        text: t("Pausa el cálculo que actualiza la luz del mapa en cada frame. Algunos efectos de luz dinámicos se ven menos realistas."),
        keys: [VOXEL],
        on: { [VOXEL]: "True" },
        note: t("Probada por Toque. Bajó la latencia de render un {render} y la GPU por frame un {gpu}.", { render: share(4.4, 1), gpu: share(6) }),
      },
      {
        id: "sky",
        control: "switch",
        title: t("Cielo gris"),
        text: t("Cambia el cielo por un fondo gris liso y quita las estrellas."),
        keys: [SKY],
        on: { [SKY]: "True" },
      },
    ],
  },
  {
    id: "world",
    title: t("Mundo"),
    hint: t("Detalle del mapa y del terreno"),
    groups: [
      {
        id: "csg",
        control: "choice",
        title: t("Detalle de piezas compuestas"),
        text: t("A qué distancia Roblox cambia las uniones y cortes por versiones con menos polígonos. Con el mínimo usa siempre la más simple."),
        keys: CSG,
        options: [
          automatic(t("Automático")),
          { id: "min", label: t("Mínimo"), values: every(CSG, "0") },
          { id: "max", label: t("Máximo"), values: every(CSG, "1000") },
        ],
        note: t("Con el mínimo, la GPU por frame apenas bajó un {gpu} en Realistic Street Soccer.", { gpu: share(1.8, 1) }),
      },
      {
        id: "grass",
        control: "switch",
        title: t("Sin hierba 3D"),
        text: t("Quita la hierba en relieve del terreno. El césped se ve plano en los mapas que la usan."),
        keys: [GRASS_MIN, GRASS_MAX],
        on: { [GRASS_MIN]: "0", [GRASS_MAX]: "0" },
        note: t("El campo de Realistic Street Soccer no usa hierba de terreno, así que ahí no cambia nada."),
      },
      {
        id: "grassMotion",
        control: "switch",
        title: t("Hierba quieta"),
        text: t("La hierba deja de moverse con el viento."),
        keys: [GRASS_MOTION],
        on: { [GRASS_MOTION]: "0" },
      },
    ],
  },
  {
    id: "screen",
    title: t("Pantalla"),
    hint: t("Ventana y escalado"),
    groups: [
      {
        id: "fullscreen",
        control: "switch",
        title: t("Pantalla completa exclusiva"),
        text: t("Con Alt+Enter, Roblox pasa a pantalla completa exclusiva en lugar de ventana sin bordes. Solo funciona con Direct3D 11."),
        keys: [ALT_ENTER],
        on: { [ALT_ENTER]: "False" },
      },
      {
        id: "dpi",
        control: "switch",
        title: t("Sin escalado DPI"),
        text: t("Roblox dibuja a la resolución real de la pantalla en lugar de seguir el zoom de Windows. Con zoom alto la interfaz se ve más pequeña."),
        keys: [DPI],
        on: { [DPI]: "True" },
      },
    ],
  },
];

const robloxDefaults: FlagValues = {
  [D3D11]: "False",
  [VULKAN]: "False",
  [OPENGL]: "False",
  [TEXTURES_ON]: "False",
  [MSAA]: "0",
  [VOXEL]: "False",
  [SKY]: "False",
  [ALT_ENTER]: "True",
  [DPI]: "False",
};

const same = (a: string | undefined, b: string | undefined) => a !== undefined && b !== undefined && a.toLowerCase() === b.toLowerCase();

function matches(group: FlagGroup, values: FlagValues, target: FlagValues) {
  return group.keys.every((k) => (k in target ? same(values[k], target[k]) : values[k] === undefined || same(values[k], robloxDefaults[k])));
}

export interface GroupState {
  selected: string | null;
  active: boolean;
  custom: string | null;
}

export function groupState(group: FlagGroup, values: FlagValues): GroupState {
  const present = group.keys.filter((k) => values[k] !== undefined);
  const custom = present.length ? present.map((k) => `${k} = ${values[k]}`).join(", ") : null;
  if (group.control === "level") {
    const raw = values[group.keys[0]];
    const level = raw === undefined ? 0 : Number(raw);
    const valid = Number.isInteger(level) && level >= 0 && level <= group.max;
    return { selected: valid ? String(level) : null, active: level > 0, custom: valid ? null : custom };
  }
  const options: { id: string; values: FlagValues }[] =
    group.control === "switch"
      ? [
          { id: "off", values: {} },
          { id: "on", values: group.on },
        ]
      : group.options;
  const hit = options.find((o) => matches(group, values, o.values));
  if (hit) return { selected: hit.id, active: hit.id !== "off" && hit.id !== "auto", custom: null };
  return { selected: null, active: true, custom };
}

export const FLAG_NAME = /^[A-Za-z][A-Za-z0-9_]{0,127}$/;

export type FlagType = "bool" | "int" | "log" | "string" | "unknown";

const typeOfPrefix: Record<string, FlagType> = { Flag: "bool", Int: "int", Log: "log", String: "string" };

export function flagType(name: string): FlagType {
  const m = /^[DS]?F(Flag|Int|Log|String)/.exec(name);
  return m ? typeOfPrefix[m[1]] : "unknown";
}

export const flagTypeLabel = (type: FlagType) => (type === "unknown" ? t("Sin prefijo") : { bool: "Bool", int: "Int", log: "Log", string: "String" }[type]);

export function validValue(type: FlagType, value: string) {
  if (type === "bool") return /^(true|false)$/i.test(value);
  if (type === "int" || type === "log") return /^-?\d+$/.test(value);
  return true;
}

export type ParsedFlags =
  | { state: "empty" }
  | { state: "invalid" }
  | { state: "ok"; values: FlagValues; count: number; skipped: number; mistyped: number };

export function parseFlagsJSON(text: string): ParsedFlags {
  if (!text.trim()) return { state: "empty" };
  let data: unknown;
  try {
    data = JSON.parse(text);
  } catch {
    return { state: "invalid" };
  }
  if (!data || typeof data !== "object" || Array.isArray(data)) return { state: "invalid" };
  const values: FlagValues = {};
  let skipped = 0;
  for (const [name, value] of Object.entries(data)) {
    if (!FLAG_NAME.test(name) || !["string", "number", "boolean"].includes(typeof value)) {
      skipped++;
      continue;
    }
    values[name] = typeof value === "boolean" ? (value ? "True" : "False") : String(value);
  }
  const names = Object.keys(values);
  return {
    state: "ok",
    values,
    count: names.length,
    skipped,
    mistyped: names.filter((n) => !validValue(flagType(n), values[n])).length,
  };
}
