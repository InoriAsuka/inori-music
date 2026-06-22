/** i18next initialisation for inori-admin. */
import i18n from "i18next";
import { initReactI18next } from "react-i18next";

const LANGUAGES = ["en", "zh", "ja"] as const;
type Lang = (typeof LANGUAGES)[number];

function localePath(lang: Lang) {
  // Admin app runs with Next.js basePath=/admin.
  return `/admin/locales/${lang}/common.json`;
}

async function loadNS(lang: Lang) {
  const res = await fetch(localePath(lang));
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
    lng: typeof window !== "undefined" ? (localStorage.getItem("inori_admin_lang") ?? "en") : "en",
    fallbackLng: "en",
    defaultNS: "common",
    interpolation: { escapeValue: false },
  });
}

export { i18n };

export function setLanguage(lang: Lang) {
  i18n.changeLanguage(lang);
  if (typeof window !== "undefined") localStorage.setItem("inori_admin_lang", lang);
}

export const SUPPORTED_LANGS: { code: Lang; label: string; nativeLabel: string }[] = [
  { code: "en", label: "English", nativeLabel: "English" },
  { code: "zh", label: "Chinese", nativeLabel: "中文" },
  { code: "ja", label: "Japanese", nativeLabel: "日本語" },
];
