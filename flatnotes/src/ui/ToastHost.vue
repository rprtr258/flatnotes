<script setup lang="ts">
import { toasts, removeToast } from "../composables/useToast";
</script>

<template>
  <Teleport to="body">
    <div class="fn-toast-host">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="fn-toast"
        :class="`fn-toast-${t.variant}`"
        @click="removeToast(t.id)"
      >
        {{ t.message }}
      </div>
    </div>
  </Teleport>
</template>

<style lang="scss" scoped>
@import "../colours";

.fn-toast-host {
  position: fixed;
  bottom: 1rem;
  right: 1rem;
  z-index: 1100;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-end;
}

.fn-toast {
  min-width: 200px;
  max-width: 350px;
  padding: 0.75rem 1rem;
  border-radius: 0.25rem;
  color: #fff;
  cursor: pointer;
  box-shadow: 0 0 10px var(--colour-shadow);
  text-align: right;
}

.fn-toast-default {
  background-color: var(--colour-text-muted);
}

.fn-toast-danger {
  background-color: #dc3545;
}

.fn-toast-success {
  background-color: #28a745;
}
</style>
