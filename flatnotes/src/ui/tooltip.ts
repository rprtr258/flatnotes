import type { Directive, DirectiveBinding } from "vue";

type TooltipEl = HTMLElement & {
  _fnTooltipShow: () => void;
  _fnTooltipHide: () => void;
  _fnTooltipKey: () => void;
  _tooltipText: string;
};

function update(el: TooltipEl, binding: DirectiveBinding<string | undefined>): void {
  const text = binding.value ?? el.getAttribute("data-tooltip") ?? "";
  el._tooltipText = text;
  el.removeAttribute("title");
  el.setAttribute("data-tooltip", text);
}

export const vTooltip: Directive<TooltipEl, string | undefined> = {
  mounted(el, binding) {
    update(el, binding);

    const show = () => {
      el.classList.add("fn-tooltip-visible");
      el.setAttribute("data-tooltip", el._tooltipText);
    };
    const hide = () => {
      el.classList.remove("fn-tooltip-visible");
    };

    el._fnTooltipShow = show;
    el._fnTooltipHide = hide;
    el.addEventListener("mouseenter", show);
    el.addEventListener("mouseleave", hide);
    el.addEventListener("focus", show);
    el.addEventListener("blur", hide);
  },
  updated(el, binding) {
    update(el, binding);
  },
  unmounted(el) {
    el.removeEventListener("mouseenter", el._fnTooltipShow);
    el.removeEventListener("mouseleave", el._fnTooltipHide);
    el.removeEventListener("focus", el._fnTooltipShow);
    el.removeEventListener("blur", el._fnTooltipHide);
  },
};
