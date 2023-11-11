import * as constants from "./constants";
import EventBus from "./eventBus";
import { getToken } from "./tokenStorage";

export default function api(path, options) {
    const {body, params, method} = options || {};

    if (params) {
      path += "?" + new URLSearchParams(params);
    }

    let fetch_options = {
      method: method || (body ? "POST" : "GET"),
      headers: (path !== "/api/token") ? {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${getToken()}`,
      } : null,
    };

    if (body) {
      fetch_options.body = JSON.stringify(body);
    }

    return fetch(path, fetch_options).catch((error) => {
      if (typeof error.response !== "undefined" && error.response.status === 401) {
        EventBus.$emit(
          "navigate",
          `${constants.basePaths.login}?${constants.params.redirect}=${encodeURI(
            window.location.pathname + window.location.search
          )}`
        );
        error.handled = true;
      }
      return Promise.reject(error);
    }).then((response) => response.json());
}
