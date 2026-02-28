import {
  startOfMonth,
  endOfMonth,
  subMonths,
  startOfYear,
  endOfYear,
  format,
} from "date-fns";
import { enUS, pl } from "date-fns/locale";
import i18next from "i18next";

function getDateLocale() {
  return i18next.language?.startsWith("pl") ? pl : enUS;
}

export function formatDate(date: string): string {
  return format(new Date(date), "PPP", { locale: getDateLocale() });
}

export function formatDateISO(date: Date): string {
  return format(date, "yyyy-MM-dd");
}

export function formatDateShort(date: Date): string {
  return format(date, "MMM d", { locale: getDateLocale() });
}

export function formatDateRange(date: Date): string {
  return format(date, "MMM d, yyyy", { locale: getDateLocale() });
}

export interface DateRange {
  from: string;
  to: string;
}

export function thisMonth(): DateRange {
  const now = new Date();
  return {
    from: formatDateISO(startOfMonth(now)),
    to: formatDateISO(endOfMonth(now)),
  };
}

export function lastMonth(): DateRange {
  const prev = subMonths(new Date(), 1);
  return {
    from: formatDateISO(startOfMonth(prev)),
    to: formatDateISO(endOfMonth(prev)),
  };
}

export function thisYear(): DateRange {
  const now = new Date();
  return {
    from: formatDateISO(startOfYear(now)),
    to: formatDateISO(endOfYear(now)),
  };
}
