import type { Money } from "@/types";
import i18next from "i18next";

export function formatMoney(m: Money): string {
  const num = parseFloat(m.amount);
  const locale = i18next.language?.startsWith("pl") ? "pl-PL" : "en-US";
  const formatted = new Intl.NumberFormat(locale, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(Math.abs(num));
  const sign = num < 0 ? "-" : "";
  return `${sign}${formatted} ${m.currency}`;
}
