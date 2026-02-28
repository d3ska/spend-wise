import { useTranslation } from "react-i18next";
import { Globe } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import axios from "axios";

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const switchLanguage = async (lang: string) => {
    await i18n.changeLanguage(lang);
    // Sync to server if authenticated (cookie-based auth)
    try {
      await axios.patch("/api/v1/auth/me", { preferred_language: lang });
    } catch {
      // Ignore — user may not be authenticated (e.g., login page)
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <Globe className="h-5 w-5" />
          <span className="sr-only">
            {i18n.language === "pl" ? "PL" : "EN"}
          </span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() => switchLanguage("en")}
          className={i18n.language === "en" ? "font-bold" : ""}
        >
          English
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => switchLanguage("pl")}
          className={i18n.language?.startsWith("pl") ? "font-bold" : ""}
        >
          Polski
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
