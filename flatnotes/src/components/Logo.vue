<script>
export default {
  props: {
    responsive: { type: Boolean, default: false },
  },
};
</script>

<style lang="scss" scoped>
@import "../colours";

@media screen and (max-width: 500px) {
  .responsive-hide {
    display: none;
  }
}

.part-icon {
  color: var(--colour-brand);
}

.part-name {
  color: var(--colour-text-muted);
}
</style>
