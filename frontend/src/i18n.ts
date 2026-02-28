import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";

import enCommon from "./locales/en/common.json";
import enErrors from "./locales/en/errors.json";
import enAuth from "./locales/en/auth.json";
import enWorkspace from "./locales/en/workspace.json";
import enTransaction from "./locales/en/transaction.json";
import enValidation from "./locales/en/validation.json";

import plCommon from "./locales/pl/common.json";
import plErrors from "./locales/pl/errors.json";
import plAuth from "./locales/pl/auth.json";
import plWorkspace from "./locales/pl/workspace.json";
import plTransaction from "./locales/pl/transaction.json";
import plValidation from "./locales/pl/validation.json";

const resources = {
  en: {
    common: enCommon,
    errors: enErrors,
    auth: enAuth,
    workspace: enWorkspace,
    transaction: enTransaction,
    validation: enValidation,
  },
  pl: {
    common: plCommon,
    errors: plErrors,
    auth: plAuth,
    workspace: plWorkspace,
    transaction: plTransaction,
    validation: plValidation,
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: "en",
    supportedLngs: ["en", "pl"],
    defaultNS: "common",
    ns: ["common", "errors", "auth", "workspace", "transaction", "validation"],
    detection: {
      order: ["localStorage", "navigator"],
      lookupLocalStorage: "i18nextLng",
      caches: ["localStorage"],
    },
    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;
