import { createContext, createElement, useContext, useState, type ReactNode } from "react";
import { thisMonth, type DateRange } from "@/lib/dates";

interface DateRangeContextValue {
  range: DateRange;
  setRange: (range: DateRange) => void;
}

const DateRangeContext = createContext<DateRangeContextValue>({
  range: thisMonth(),
  setRange: () => {},
});

export function DateRangeProvider({ children }: { children: ReactNode }) {
  const [range, setRange] = useState<DateRange>(thisMonth);

  return createElement(
    DateRangeContext.Provider,
    { value: { range, setRange } },
    children,
  );
}

export function useDateRange() {
  return useContext(DateRangeContext);
}
