import { reactive } from "vue";

type OkVariant = "primary" | "danger" | "warning";

type ConfirmState = {
  open: boolean;
  message: string;
  title: string;
  okVariant: OkVariant;
  resolve: ((value: boolean) => void) | null;
};

export const confirmState = reactive<ConfirmState>({
  open: false,
  message: "",
  title: "",
  okVariant: "primary",
  resolve: null,
});

export function msgBoxConfirm(
  message: string,
  opts: { title?: string; okVariant?: OkVariant; centered?: boolean } = {},
): Promise<boolean> {
  return new Promise((resolve) => {
    confirmState.message = message;
    confirmState.title = opts.title ?? "Confirm";
    confirmState.okVariant = opts.okVariant ?? "primary";
    confirmState.resolve = resolve;
    confirmState.open = true;
  });
}

export function resolveConfirm(value: boolean): void {
  confirmState.resolve?.(value);
  confirmState.resolve = null;
  confirmState.open = false;
}
