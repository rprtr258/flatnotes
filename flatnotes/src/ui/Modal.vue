<script setup lang="ts">
import { watch, nextTick, ref } from "vue";

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    title?: string;
    centered?: boolean;
    hideHeader?: boolean;
    hideFooter?: boolean;
    okTitle?: string;
    cancelTitle?: string;
    okVariant?: "primary" | "danger" | "warning";
  }>(),
  {
    title: "",
    centered: true,
    hideHeader: false,
    hideFooter: false,
    okTitle: "OK",
    cancelTitle: "Cancel",
    okVariant: "primary",
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  ok: [];
  cancel: [];
}>();

const dialogEl = ref<HTMLElement | null>(null);

watch(
  () => props.modelValue,
  async (open) => {
    if (open) {
      document.body.classList.add("modal-open");
      await nextTick();
      // focus the dialog for keyboard escape
      dialogEl.value?.focus();
    } else {
      document.body.classList.remove("modal-open");
    }
  },
);

function close(): void {
  emit("update:modelValue", false);
}

function onOk(): void {
  emit("ok");
  close();
}

function onCancel(): void {
  emit("cancel");
  close();
}

function onBackdrop(): void {
  close();
}
</script>

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fn-modal-backdrop" @click.self="onBackdrop">
      <div
        ref="dialogEl"
        class="fn-modal-dialog"
        :class="{ 'fn-modal-centered': centered }"
        tabindex="-1"
        @keyup.esc="onCancel"
      >
        <div class="fn-modal-content">
          <div v-if="!hideHeader" class="fn-modal-header">
            <h5 class="fn-modal-title">{{ title }}</h5>
            <button type="button" class="fn-modal-close" aria-label="Close" @click="onCancel">
              <i class="bi bi-x"></i>
            </button>
          </div>
          <div class="fn-modal-body">
            <slot />
          </div>
          <div v-if="!hideFooter" class="fn-modal-footer">
            <slot name="footer">
              <button type="button" class="btn btn-secondary" @click="onCancel">{{ cancelTitle }}</button>
              <button
                type="button"
                class="btn"
                :class="okVariant === 'danger' ? 'btn-danger' : okVariant === 'warning' ? 'btn-warning' : 'btn-primary'"
                @click="onOk"
              >
                {{ okTitle }}
              </button>
            </slot>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style lang="scss" scoped>
@import "../colours";

.fn-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1050;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  padding: 10px;
}

.fn-modal-dialog {
  position: relative;
  margin: auto;
  width: 100%;
  max-width: 500px;
  outline: none;
}

.fn-modal-centered {
  align-self: center;
}

.fn-modal-content {
  background-color: var(--colour-background-elevated);
  border: 1px solid var(--colour-border);
  border-radius: 0.3rem;
  display: flex;
  flex-direction: column;
  width: 100%;
  pointer-events: auto;
}

.fn-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid var(--colour-border);
}

.fn-modal-title {
  margin: 0;
  font-size: 1.25rem;
  color: var(--colour-text);
}

.fn-modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  line-height: 1;
  color: var(--colour-text);
  cursor: pointer;
  padding: 0 0.25rem;
}

.fn-modal-body {
  padding: 1rem;
  color: var(--colour-text);
}

.fn-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem;
  border-top: 1px solid var(--colour-border);
}
</style>
