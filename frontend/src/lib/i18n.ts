import { useSyncExternalStore } from "react";
import corePt from "@/locales/core/pt.json";
import coreEn from "@/locales/core/en.json";
import uiPt from "@/locales/ui/pt.json";
import uiEn from "@/locales/ui/en.json";

export type Lang = "es" | "pt" | "en";

const dictionaries: Record<Lang, Record<string, string>> = {
  es: {},
  pt: { ...corePt, ...uiPt },
  en: { ...coreEn, ...uiEn },
};

const locales: Record<Lang, string> = { es: "es-CO", pt: "pt-BR", en: "en-US" };

let lang: Lang = "es";
const listeners = new Set<() => void>();

export function setLang(next: Lang) {
  if (next === lang) return;
  lang = next;
  document.documentElement.lang = next;
  listeners.forEach((l) => l());
}

export function currentLang() {
  return lang;
}

export function locale() {
  return locales[lang];
}

export function useLang() {
  return useSyncExternalStore(
    (l) => {
      listeners.add(l);
      return () => listeners.delete(l);
    },
    currentLang,
  );
}

export function t(text: string, params?: Record<string, string | number>) {
  const out = dictionaries[lang][text] || text;
  if (!params) return out;
  return out.replace(/\{(\w+)\}/g, (hole, name: string) => (name in params ? String(params[name]) : hole));
}

export function mark(text: string) {
  return text;
}
