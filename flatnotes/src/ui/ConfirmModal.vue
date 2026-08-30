<script setup lang="ts">
import Modal from "./Modal.vue";
import { confirmState, resolveConfirm } from "../composables/useModal";

function onOk(): void {
  resolveConfirm(true);
}

function onCancel(): void {
  resolveConfirm(false);
}
</script>

<template>
  <Modal
    :model-value="confirmState.open"
    :title="confirmState.title"
    :ok-title="'OK'"
    :ok-variant="confirmState.okVariant"
    @ok="onOk"
    @cancel="onCancel"
    @update:model-value="(v: boolean) => { if (!v) onCancel(); }"
  >
    {{ confirmState.message }}
  </Modal>
</template>
