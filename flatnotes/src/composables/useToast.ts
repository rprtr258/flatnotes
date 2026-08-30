import { reactive } from "vue";

export type ToastVariant = "default" | "danger" | "success";

export interface ToastItem {
  id: number;
  message: string;
  variant: ToastVariant;
}

export const toasts = reactive<ToastItem[]>([]);

let nextId = 1;

export function toast(message: string, opts: { variant?: ToastVariant; timeout?: number } = {}): void {
  const id = nextId++;
  toasts.push({ id, message, variant: opts.variant ?? "default" });
  const timeout = opts.timeout ?? 4000;
  if (timeout > 0) {
    setTimeout(() => removeToast(id), timeout);
  }
}

export function removeToast(id: number): void {
  const i = toasts.findIndex((t) => t.id === id);
  if (i !== -1) toasts.splice(i, 1);
}
