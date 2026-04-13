import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { CustomMenuItem } from "@/types";
import { translateMessage } from "@/i18n";
import { resolveCustomPageMenuItem } from "@/utils/customMenu";

interface ResolveRouteDocumentTitleOptions {
  siteName?: string;
  publicCustomMenuItems?: CustomMenuItem[];
  adminCustomMenuItems?: CustomMenuItem[];
  isAdmin?: boolean;
}

function resolveSiteName(siteName?: string): string {
  return typeof siteName === "string" && siteName.trim()
    ? siteName.trim()
    : "Sub2API";
}

/**
 * 统一生成页面标题，避免多处写入 document.title 产生覆盖冲突。
 * 优先使用 titleKey 通过 i18n 翻译，fallback 到静态 routeTitle。
 */
export function resolveDocumentTitle(
  routeTitle: unknown,
  siteName?: string,
  titleKey?: string,
): string {
  const normalizedSiteName = resolveSiteName(siteName);

  if (typeof titleKey === "string" && titleKey.trim()) {
    const translated = translateMessage(titleKey);
    if (translated && translated !== titleKey) {
      return `${translated} - ${normalizedSiteName}`;
    }
  }

  if (typeof routeTitle === "string" && routeTitle.trim()) {
    return `${routeTitle.trim()} - ${normalizedSiteName}`;
  }

  return normalizedSiteName;
}

export function resolveRouteDocumentTitle(
  route: Pick<RouteLocationNormalizedLoaded, "name" | "meta" | "params">,
  options: ResolveRouteDocumentTitleOptions = {},
): string {
  const normalizedSiteName = resolveSiteName(options.siteName);

  if (route.name === "CustomPage") {
    const menuItemId =
      typeof route.params.id === "string" ? route.params.id : "";
    const menuItem = resolveCustomPageMenuItem(
      menuItemId,
      options.publicCustomMenuItems ?? [],
      options.adminCustomMenuItems ?? [],
      options.isAdmin === true,
    );

    if (menuItem?.label.trim()) {
      return `${menuItem.label.trim()} - ${normalizedSiteName}`;
    }
  }

  const titleKey =
    typeof route.meta.titleKey === "string" ? route.meta.titleKey : undefined;
  return resolveDocumentTitle(route.meta.title, normalizedSiteName, titleKey);
}
