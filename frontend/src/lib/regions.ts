import type { Region } from "./types";

export const PRIVATE = "private";

const countries = new Intl.DisplayNames(["es"], { type: "region" });

export function countryName(code: string) {
  try {
    return countries.of(code) ?? code;
  } catch {
    return code;
  }
}

export function regionCity(key: string) {
  return key.slice(key.indexOf("/") + 1);
}

export function regionPlace(r: Region) {
  return r.area ? `${r.area}, ${countryName(r.country)}` : countryName(r.country);
}

export function bestRegion(regions: Region[]) {
  return regions.find((r) => r.rttMs > 0) ?? null;
}
