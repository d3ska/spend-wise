import { NavLink } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  AlertTriangle,
  LayoutDashboard,
  Receipt,
  Tags,
  Wallet,
  Cog,
  Sparkles,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useListConnections } from "@/api/bank-connections";

export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const { t } = useTranslation(["common", "workspace"]);
  const { data: connections } = useListConnections();
  const hasExpiringConnection = connections?.some(
    (c) => c.expiry_status === "warning",
  );

  const links = [
    { to: "/dashboard", label: t("common:nav.dashboard"), icon: LayoutDashboard },
    { to: "/transactions", label: t("common:nav.transactions"), icon: Receipt },
    { to: "/categories", label: t("common:nav.categories"), icon: Tags },
    { to: "/budgets", label: t("common:nav.budgets"), icon: Wallet },
    { to: "/rules", label: t("common:nav.rules"), icon: Sparkles },
    { to: "/settings", label: t("common:nav.settings"), icon: Cog },
  ];

  return (
    <nav className="flex flex-col gap-0.5 px-3 py-4">
      {hasExpiringConnection && (
        <NavLink
          to="/settings"
          onClick={onNavigate}
          className="flex items-center gap-2 rounded-md bg-yellow-50 border border-yellow-200 px-3 py-2 mb-2 text-sm text-yellow-800 hover:bg-yellow-100 transition-colors dark:bg-yellow-950 dark:border-yellow-800 dark:text-yellow-200"
        >
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span>{t("workspace:bank.expiringWarning")}</span>
        </NavLink>
      )}
      {links.map(({ to, label, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors",
              isActive
                ? "bg-accent text-accent-foreground font-semibold"
                : "text-muted-foreground hover:bg-accent/50 hover:text-foreground font-medium",
            )
          }
        >
          <Icon className="h-4 w-4" />
          {label}
        </NavLink>
      ))}
    </nav>
  );
}
