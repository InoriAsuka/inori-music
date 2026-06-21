/**
 * lib/i18n.ts — i18next initialisation for inori-web.
 * Loads locale files from /public/locales/{lang}/common.json.
 * Call initI18n() once at the app root; it is idempotent.
 */
import i18n from "i18next";
import { initReactI18next } from "react-i18next";

const LANGUAGES = ["en", "zh", "ja"] as const;
type Lang = (typeof LANGUAGES)[number];

async function loadNS(lang: Lang) {
  const res = await fetch(`/locales/${lang}/common.json`);
  return res.json();
}

let initialized = false;

export async function initI18n() {
  if (initialized) return;
  initialized = true;

  const resources: Record<string, { common: object }> = {};
  await Promise.all(
    LANGUAGES.map(async (lang) => {
      resources[lang] = { common: await loadNS(lang) };
    })
  );

  await i18n.use(initReactI18next).init({
    resources,
    lng: typeof window !== "undefined" ? (localStorage.getItem("inori_lang") ?? "en") : "en",
    fallbackLng: "en",
    defaultNS: "common",
    interpolation: { escapeValue: false },
  });
}

export { i18n };

export function setLanguage(lang: Lang) {
  i18n.changeLanguage(lang);
  if (typeof window !== "undefined") localStorage.setItem("inori_lang", lang);
}

export const SUPPORTED_LANGS: { code: Lang; label: string; nativeLabel: string }[] = [
  { code: "en", label: "English", nativeLabel: "English" },
  { code: "zh", label: "Chinese", nativeLabel: "中文" },
  { code: "ja", label: "Japanese", nativeLabel: "日本語" },
];
