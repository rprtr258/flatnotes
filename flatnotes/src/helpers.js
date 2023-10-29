export function getSearchParam(paramName, defaultValue = null) {
  let urlSearchParams = new URLSearchParams(window.location.search);
  let paramValue = urlSearchParams.get(paramName);
  return (paramValue != null) ? paramValue : defaultValue;
}

export function getSearchParamBool(paramName, defaultValue = null) {
  let paramValue = getSearchParam(paramName)
  if (paramValue == null) {
    return defaultValue
  }
  let paramValueLowerCase = paramValue.toLowerCase();
  switch (paramValueLowerCase ) {
    case "true":
      return true
    case "false":
      return false
    default:
      return defaultValue;
  }
}

export function getSearchParamInt(paramName, defaultValue = null) {
  let paramValue = getSearchParam(paramName)
  if (paramValue == null) {
    return defaultValue
  }

  let paramValueInt = parseInt(paramValue);
  return !isNaN(paramValueInt) ? paramValueInt : defaultValue;
}

export function setSearchParam(paramName, value) {
  let url = new URL(window.location.href);
  let urlSearchParams = new URLSearchParams(url.search);
  urlSearchParams.set(paramName, value);
  url.search = urlSearchParams.toString();
  window.history.replaceState({}, "", url.toString());
}
