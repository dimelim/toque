import { useEffect, useRef, useState } from "react";
import type { KeyboardEvent as ReactKeyboardEvent, PointerEvent as ReactPointerEvent } from "react";
import { cx } from "@/utils/cx";
import { t } from "@/lib/i18n";
import type { HudConfig, HudLayout } from "@/lib/types";

const margin = 32;
const snapRange = 48;
const sendEvery = 40;

function snap(v: number, max: number) {
  if (Math.abs(v - margin) < snapRange) return margin;
  if (Math.abs(v - (max - margin)) < snapRange) return max - margin;
  if (Math.abs(v - max / 2) < snapRange) return max / 2;
  return v;
}

function MiniHud({ cfg }: { cfg: HudConfig }) {
  return (
    <svg viewBox="0 0 120 84" className="size-full" preserveAspectRatio="none" aria-hidden>
      {cfg.showPing ? <rect x="10" y="8" width="44" height="7" rx="3.5" className="fill-volt" /> : null}
      {cfg.showKeys ? (
        <g className="fill-ink-3">
          <rect x="34" y="22" width="20" height="17" rx="4" className="fill-volt" />
          <rect x="10" y="43" width="20" height="17" rx="4" />
          <rect x="34" y="43" width="20" height="17" rx="4" className="fill-volt" />
          <rect x="58" y="43" width="20" height="17" rx="4" />
          <rect x="10" y="64" width="32" height="13" rx="4" />
          <rect x="46" y="64" width="32" height="13" rx="4" />
        </g>
      ) : null}
      {cfg.showMouse ? <rect x="86" y="30" width="24" height="40" rx="11" className="fill-ink-4" /> : null}
    </svg>
  );
}

export function HudEditor({ cfg, layout, disabled, onPlace }: { cfg: HudConfig; layout: HudLayout; disabled: boolean; onPlace: (x: number, y: number, save: boolean) => void }) {
  const frame = useRef<HTMLDivElement>(null);
  const grab = useRef<{ dx: number; dy: number; sent: number; x: number; y: number } | null>(null);
  const [pos, setPos] = useState({ x: cfg.x, y: cfg.y });
  const [dragging, setDragging] = useState(false);
  useEffect(() => {
    if (!grab.current) setPos({ x: cfg.x, y: cfg.y });
  }, [cfg.x, cfg.y]);

  const maxX = Math.max(0, layout.screenW - layout.panelW);
  const maxY = Math.max(0, layout.screenH - layout.panelH);

  const toScreen = (e: ReactPointerEvent) => {
    const r = frame.current!.getBoundingClientRect();
    const k = layout.screenW / r.width;
    return { x: (e.clientX - r.left) * k, y: (e.clientY - r.top) * k };
  };

  const down = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (disabled) return;
    e.currentTarget.setPointerCapture(e.pointerId);
    const p = toScreen(e);
    grab.current = { dx: p.x - pos.x, dy: p.y - pos.y, sent: 0, x: pos.x, y: pos.y };
    setDragging(true);
  };

  const move = (e: ReactPointerEvent<HTMLDivElement>) => {
    const g = grab.current;
    if (!g) return;
    const p = toScreen(e);
    const x = Math.round(snap(Math.min(maxX, Math.max(0, p.x - g.dx)), maxX));
    const y = Math.round(snap(Math.min(maxY, Math.max(0, p.y - g.dy)), maxY));
    g.x = x;
    g.y = y;
    setPos({ x, y });
    const now = performance.now();
    if (now - g.sent > sendEvery) {
      g.sent = now;
      onPlace(x, y, false);
    }
  };

  const up = () => {
    const g = grab.current;
    if (!g) return;
    grab.current = null;
    setDragging(false);
    onPlace(g.x, g.y, true);
  };

  const nudge = (e: ReactKeyboardEvent<HTMLDivElement>) => {
    const step = e.shiftKey ? 40 : 8;
    const moves: Record<string, [number, number]> = { ArrowLeft: [-step, 0], ArrowRight: [step, 0], ArrowUp: [0, -step], ArrowDown: [0, step] };
    const m = moves[e.key];
    if (disabled || !m) return;
    e.preventDefault();
    const x = Math.min(maxX, Math.max(0, pos.x + m[0]));
    const y = Math.min(maxY, Math.max(0, pos.y + m[1]));
    setPos({ x, y });
    onPlace(x, y, false);
  };

  const save = (e: ReactKeyboardEvent<HTMLDivElement>) => {
    if (!disabled && e.key.startsWith("Arrow")) onPlace(pos.x, pos.y, true);
  };

  const pct = (v: number, of: number) => `${(v / of) * 100}%`;
  return (
    <div>
      <div
        ref={frame}
        className="relative w-full overflow-hidden rounded-xl bg-surface-1 shadow-[inset_0_0_0_1px_var(--color-line-strong)]"
        style={{ aspectRatio: `${layout.screenW} / ${layout.screenH}` }}
      >
        <div className="absolute inset-x-0 bottom-0 h-[3.5%] bg-surface-2" />
        <div
          role="slider"
          aria-label={t("Posición del HUD")}
          aria-valuetext={`x ${pos.x}, y ${pos.y}`}
          onPointerDown={down}
          onPointerMove={move}
          onPointerUp={up}
          onPointerCancel={up}
          onKeyDown={nudge}
          onKeyUp={save}
          aria-valuemin={0}
          aria-valuemax={maxX}
          aria-valuenow={pos.x}
          tabIndex={disabled ? -1 : 0}
          className={cx(
            "absolute touch-none rounded-[10%] bg-surface-3 p-[1.2%] ring-1 transition-[box-shadow,background-color] duration-150 outline-none select-none focus-visible:ring-2 focus-visible:ring-border-focus-ring",
            disabled ? "cursor-not-allowed opacity-60 ring-line-strong" : dragging ? "cursor-grabbing bg-surface-4 ring-volt" : "cursor-grab ring-volt-line hover:ring-volt",
          )}
          style={{ left: pct(pos.x, layout.screenW), top: pct(pos.y, layout.screenH), width: pct(layout.panelW, layout.screenW), height: pct(layout.panelH, layout.screenH) }}
        >
          <MiniHud cfg={cfg} />
        </div>
      </div>
      <div className="mt-2 flex items-center justify-between text-caption-1-medium text-ink-3">
        <span>
          {t("Tu pantalla {w} × {h}", { w: layout.screenW, h: layout.screenH })}
        </span>
        <span className="num">
          X {pos.x} · Y {pos.y}
        </span>
      </div>
    </div>
  );
}
