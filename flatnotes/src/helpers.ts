export function getSearchParam(paramName: string, defaultValue: string | null = null): string | null {
  const urlSearchParams = new URLSearchParams(window.location.search);
  const paramValue = urlSearchParams.get(paramName);
  return paramValue != null ? paramValue : defaultValue;
}

export function getSearchParamBool(paramName: string, defaultValue: boolean | null = null): boolean | null {
  const paramValue = getSearchParam(paramName);
  if (paramValue == null) {
    return defaultValue;
  }
  switch (paramValue.toLowerCase()) {
    case "true":
      return true;
    case "false":
      return false;
    default:
      return defaultValue;
  }
}

export function getSearchParamInt(paramName: string, defaultValue: number | null = null): number | null {
  const paramValue = getSearchParam(paramName);
  if (paramValue == null) {
    return defaultValue;
  }
  const paramValueInt = parseInt(paramValue, 10);
  return !isNaN(paramValueInt) ? paramValueInt : defaultValue;
}

export function setSearchParam(paramName: string, value: string): void {
  const url = new URL(window.location.href);
  const urlSearchParams = new URLSearchParams(url.search);
  urlSearchParams.set(paramName, value);
  url.search = urlSearchParams.toString();
  window.history.replaceState({}, "", url.toString());
}
