import type { TFunction } from "i18next";

/**
 * Returns the display name for a category, using i18n translation for default
 * categories (those with a slug) and the raw name for custom categories.
 */
export function getCategoryDisplayName(
  cat: { name: string; slug: string | null },
  t: TFunction,
): string {
  if (cat.slug) {
    return t(`transaction:categoryNames.${cat.slug}`, {
      defaultValue: cat.name,
    });
  }
  return cat.name;
}

/**
 * Returns true if the category is the "Uncategorized" default category.
 */
export function isUncategorized(cat: {
  slug: string | null;
  name: string;
}): boolean {
  return cat.slug === "uncategorized";
}
