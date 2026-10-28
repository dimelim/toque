import { locale } from "./i18n";
import type { Region } from "./types";

export const PRIVATE = "private";

let countries: { tag: string; names: Intl.DisplayNames } | null = null;

export function countryName(code: string) {
  try {
    const tag = locale();
    if (countries?.tag !== tag) countries = { tag, names: new Intl.DisplayNames([tag], { type: "region" }) };
    return countries.names.of(code) ?? code;
  } catch {
    return code;
  }
}

export function regionCity(key: string) {
  return key.slice(key.indexOf("/") + 1);
}

export function regionCountry(key: string) {
  return key.slice(0, Math.max(0, key.indexOf("/")));
}

export function regionPlace(r: Region) {
  return r.area ? `${r.area}, ${countryName(r.country)}` : countryName(r.country);
}

export function bestRegion(regions: Region[]) {
  return regions.find((r) => r.rttMs > 0) ?? null;
}
