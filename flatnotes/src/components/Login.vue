<script setup lang="ts">
import { ref, watch, onMounted } from "vue";

import * as constants from "../constants";
import * as helpers from "../helpers";
import { eventBus } from "../eventBus";
import Logo from "./Logo.vue";
import Icon from "../ui/Icon.vue";
import api from "../api";
import { setToken } from "../tokenStorage";
import { toast } from "../composables/useToast";
import type { TokenResponse } from "../types";

const props = defineProps<{ authType: string | null }>();

const usernameInput = ref<string | null>(null);
const passwordInput = ref<string | null>(null);
const totpInput = ref<string | null>(null);
const rememberMeInput = ref(false);

function skipIfNoneAuthType(): void {
  if (props.authType === constants.authTypes.none) {
    eventBus.emit("navigate", { href: constants.basePaths.home });
  }
}

watch(
  () => props.authType,
  () => skipIfNoneAuthType(),
);

function login(): void {
  api<TokenResponse>("/api/token", {
    body: {
      username: usernameInput.value,
      password: (passwordInput.value ?? "") + (props.authType === constants.authTypes.totp ? (totpInput.value ?? "") : ""),
    },
  })
    .then((response) => {
      setToken(response.access_token, rememberMeInput.value);
      const redirectPath = helpers.getSearchParam(constants.params.redirect);
      eventBus.emit("navigate", { href: redirectPath || constants.basePaths.home });
    })
    .catch((error) => {
      const err = error as { handled?: boolean; status?: number };
      if (err.handled) {
        return;
      } else if (err.status != null && [400, 422].includes(err.status)) {
        toast("Incorrect login credentials ✘", { variant: "danger" });
      } else {
        eventBus.emit("unhandled-server-error", { error });
      }
    })
    .finally(() => {
      usernameInput.value = null;
      passwordInput.value = null;
      totpInput.value = null;
      rememberMeInput.value = false;
    });
}

onMounted(skipIfNoneAuthType);
</script>

<template>
  <div class="d-flex flex-column justify-content-center align-items-center">
    <!-- Logo -->
    <Logo class="mb-5" />
    <div
      v-if="authType != null && authType !== constants.authTypes.none"
      class="d-flex flex-column justify-content-center align-items-center"
    >
      <form
        v-show="authType != null"
        class="login-form d-flex flex-column align-items-center"
        @submit.prevent="login"
      >
        <div class="mb-1">
          <!-- Username -->
          <div class="mb-1">
            <input
              type="text"
              placeholder="Username"
              class="form-control"
              id="username"
              autocomplete="username"
              v-model="usernameInput"
              autofocus
              required
            />
          </div>

          <!-- Password -->
          <div class="mb-1">
            <input
              type="password"
              placeholder="Password"
              class="form-control"
              id="password"
              autocomplete="current-password"
              v-model="passwordInput"
              required
            />
          </div>

          <!-- 2FA -->
          <div v-if="authType === constants.authTypes.totp" class="mb-1">
            <input
              type="text"
              inputmode="numeric"
              pattern="[0-9]*"
              placeholder="2FA Code"
              class="form-control"
              id="totp"
              autocomplete="one-time-code"
              v-model="totpInput"
              required
            />
          </div>
        </div>

        <!-- Remember Me -->
        <div class="mb-3 form-check">
          <input
            type="checkbox"
            class="form-check-input"
            id="rememberMe"
            v-model="rememberMeInput"
          />
          <label class="form-check-label" for="rememberMe">Remember Me</label>
        </div>

        <!-- Button -->
        <button type="submit" class="bttn">
          <Icon name="box-arrow-in-right" /> Log In
        </button>
      </form>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.login-form {
  input {
    color: var(--colour-text);
    background-color: var(--colour-background-elevated);
    border-color: var(--colour-border);
  }
}
</style>
