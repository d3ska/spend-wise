import axios from "axios";
import i18next from "i18next";

/**
 * Extract a user-friendly error message from an API error.
 *
 * - If the response has a `code` field, look up a translated message
 * - 403 → permission message from translations
 * - Falls back to `message` field, then provided fallback
 */
export function getErrorMessage(err: unknown, fallback: string): string {
  if (!axios.isAxiosError(err) || !err.response) {
    return fallback;
  }

  const { status, data } = err.response;

  if (status === 403) {
    return i18next.t("errors:permissionDenied");
  }

  // Use error code for translated message
  if (typeof data?.code === "string" && data.code.length > 0) {
    const translated = i18next.t(`errors:${data.code}`);
    // If the translation key exists (not the same as the key), use it
    if (translated !== data.code) {
      return translated;
    }
  }

  // Fall back to message field
  if (typeof data?.message === "string" && data.message.length > 0) {
    return data.message;
  }

  return fallback;
}
