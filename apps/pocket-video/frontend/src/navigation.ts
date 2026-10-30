// Spatial navigation shared by keyboard and gamepad. Pages do not need to opt
// in: anything focusable that is visible participates in the focus ring.
const FOCUSABLE = 'button:not([disabled]),a[href],[tabindex="0"],input:not([disabled]),select:not([disabled]),video[controls],audio[controls]';

function visibleItems(): HTMLElement[] {
  return [...document.querySelectorAll<HTMLElement>(FOCUSABLE)].filter((element) => element.getClientRects().length > 0);
}

function focusStep(delta: number): void {
  const items = visibleItems();
  if (items.length === 0) return;
  const index = items.indexOf(document.activeElement as HTMLElement);
  const next = (Math.max(index, -1) + delta + items.length) % items.length;
  items[next]?.focus();
}

function isTextInput(target: EventTarget | null): boolean {
  return target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement;
}

export function installNavigation(onBack: () => void, onConfirm: () => void): () => void {
  const keydown = (event: KeyboardEvent) => {
    const target = event.target as HTMLElement | null;
    if (event.key === 'Escape') {
      onBack();
      event.preventDefault();
      return;
    }
    if (event.key === 'Enter') {
      if (target && target !== document.body && (target instanceof HTMLButtonElement || target instanceof HTMLAnchorElement || target instanceof HTMLSelectElement)) {
        target.click();
        event.preventDefault();
        return;
      }
      if (target instanceof HTMLInputElement && target.type === 'search') return;
      onConfirm();
      event.preventDefault();
      return;
    }
    if (!['ArrowDown', 'ArrowUp', 'ArrowLeft', 'ArrowRight'].includes(event.key)) return;
    if (isTextInput(target)) return;
    focusStep(event.key === 'ArrowDown' || event.key === 'ArrowRight' ? 1 : -1);
    event.preventDefault();
  };
  window.addEventListener('keydown', keydown);

  let frame = 0;
  let lastMove = 0;
  const pressed = new Map<number, boolean[]>();
  const poll = () => {
    for (const pad of navigator.getGamepads?.() ?? []) {
      if (!pad) continue;
      const previous = pressed.get(pad.index) ?? [];
      const rising = (button: number) => Boolean(pad.buttons[button]?.pressed && !previous[button]);
      pressed.set(pad.index, pad.buttons.map((button) => button.pressed));
      if (rising(0)) {
        const active = document.activeElement;
        if (active instanceof HTMLElement && active !== document.body) active.click();
        else onConfirm();
      }
      if (rising(1)) onBack();
      const axis = Math.abs(pad.axes[0] ?? 0) > Math.abs(pad.axes[1] ?? 0) ? pad.axes[0] : pad.axes[1];
      const now = performance.now();
      if (Math.abs(axis ?? 0) > 0.65 && now - lastMove > 220) {
        focusStep((axis ?? 0) > 0 ? 1 : -1);
        lastMove = now;
      }
    }
    frame = requestAnimationFrame(poll);
  };
  frame = requestAnimationFrame(poll);

  return () => {
    window.removeEventListener('keydown', keydown);
    cancelAnimationFrame(frame);
  };
}
