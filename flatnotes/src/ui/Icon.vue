<script setup lang="ts">
defineProps<{ name: string }>();
</script>

<template>
  <i class="bi" :class="`bi-${name}`" aria-hidden="true"></i>
</template>
