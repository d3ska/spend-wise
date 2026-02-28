import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { CalendarIcon } from "lucide-react";
import { useState } from "react";
import { format } from "date-fns";
import { formatDateISO, thisMonth, lastMonth, thisYear } from "@/lib/dates";
import type { DateRange as DateRangeType } from "@/lib/dates";
import type { DateRange as RDPDateRange } from "react-day-picker";

interface Props {
  value: DateRangeType;
  onChange: (range: DateRangeType) => void;
}

export function DateRangePicker({ value, onChange }: Props) {
  const { t } = useTranslation(["transaction"]);
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<RDPDateRange | undefined>(undefined);

  const label = `${format(new Date(value.from), "MMM d")} - ${format(new Date(value.to), "MMM d, yyyy")}`;

  const presets = [
    { label: t("transaction:dateRange.thisMonth"), fn: thisMonth },
    { label: t("transaction:dateRange.lastMonth"), fn: lastMonth },
    { label: t("transaction:dateRange.thisYear"), fn: thisYear },
  ];

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (!next) {
      setDraft(undefined);
    }
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <CalendarIcon className="h-4 w-4" />
          {label}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="flex gap-2 border-b p-2">
          {presets.map((p) => (
            <Button
              key={p.label}
              variant="ghost"
              size="sm"
              onClick={() => {
                onChange(p.fn());
                setOpen(false);
                setDraft(undefined);
              }}
            >
              {p.label}
            </Button>
          ))}
        </div>
        <Calendar
          mode="range"
          selected={draft}
          onSelect={(range) => {
            const isSecondClick = draft?.from != null;
            if (isSecondClick && range?.from && range?.to) {
              // Second click on a different date — commit the range.
              onChange({
                from: formatDateISO(range.from),
                to: formatDateISO(range.to),
              });
              setOpen(false);
              setDraft(undefined);
            } else if (isSecondClick && (!range?.from || !range?.to)) {
              // Second click on the same date — RDP deselects it.
              // Commit a single-day range.
              const day = draft.from!;
              onChange({ from: formatDateISO(day), to: formatDateISO(day) });
              setOpen(false);
              setDraft(undefined);
            } else {
              setDraft(range);
            }
          }}
          numberOfMonths={2}
        />
      </PopoverContent>
    </Popover>
  );
}
