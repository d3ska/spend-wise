import { useTranslation } from "react-i18next";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { initials } from "@/lib/format";

interface MemberAvatarProps {
  displayName: string;
  avatarUrl: string;
  isCurrentUser?: boolean;
}

export function MemberAvatar({ displayName, avatarUrl, isCurrentUser }: MemberAvatarProps) {
  const { t } = useTranslation(["common"]);
  const tooltip = isCurrentUser ? `${displayName} ${t("common:youSuffix")}` : displayName;

  return (
    <Avatar className="h-6 w-6 shrink-0" title={tooltip}>
      <AvatarImage src={avatarUrl} />
      <AvatarFallback className="text-[10px]">{initials(displayName)}</AvatarFallback>
    </Avatar>
  );
}
