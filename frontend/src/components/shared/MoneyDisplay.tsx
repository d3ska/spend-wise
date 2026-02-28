import type { Money } from "@/types";
import { formatMoney } from "@/lib/money";
import { cn } from "@/lib/utils";

export function MoneyDisplay({
  value,
  className,
}: {
  value: Money;
  className?: string;
}) {
  const num = parseFloat(value.amount);
  return (
    <span
      className={cn(
        "tabular-nums",
        num < 0 && "text-destructive",
        className,
      )}
    >
      {formatMoney(value)}
    </span>
  );
}
