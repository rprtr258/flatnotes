<script>
import * as constants from "../constants";
import * as helpers from "../helpers";
import EventBus from "../eventBus";
import Logo from "./Logo";
import api from "../api";

export default {
  data() {
    return {
      usernameInput: null,
      passwordInput: null,
      totpInput: null,
      rememberMeInput: false,
    };
  },

  watch: {
    authType() {
      this.skipIfNoneAuthType();
    },
  },

  methods: {
    skipIfNoneAuthType() {
      // Skip past the login page if authentication is disabled
      if (this.authType == constants.authTypes.none) {
        EventBus.$emit("navigate", constants.basePaths.home);
      }
    },

    login() {
      let parent = this;
      api("/api/token", {
        body: {
          username: this.usernameInput,
          password: this.passwordInput + (this.authType == constants.authTypes.totp ? this.totpInput : ""),
        },
      })
        .then((response) => {
          sessionStorage.setItem("token", response.access_token);
          if (parent.rememberMeInput == true) {
            localStorage.setItem("token", response.access_token);
          }
          let redirectPath = helpers.getSearchParam(constants.params.redirect);
          EventBus.$emit("navigate", redirectPath || constants.basePaths.home);
        })
        .catch((error) => {
          if (error.handled) {
            return;
          } else if (typeof error.response !== "undefined" && [400, 422].includes(error.response.status)) {
            parent.$bvToast.toast("Incorrect login credentials ✘", {
              variant: "danger",
              noCloseButton: true,
              toaster: "b-toaster-bottom-right",
            });
          } else {
            EventBus.$emit("unhandledServerError", error);
          }
        })
        .finally(() => {
          parent.usernameInput = null;
          parent.passwordInput = null;
          parent.totpInput = null;
          parent.rememberMeInput = false;
        });
    },
  },

  created() {
    this.constants = constants;
    this.skipIfNoneAuthType();
  },
};
</script>

<style lang="scss" scoped>
.login-form {
  input {
    color: var(--colour-text);
    background-color: var(--colour-background-elevated);
    border-color: var(--colour-border);
  }
}
</style>
