import type { ComponentType } from "react";
import { cx } from "@/utils/cx";

export type TileTone = "lime" | "cyan" | "amber" | "rose";

const tones: Record<TileTone, string> = {
  lime: "from-tile-lime-from to-tile-lime-to text-volt-ink",
  cyan: "from-tile-cyan-from to-tile-cyan-to text-canvas",
  amber: "from-tile-amber-from to-tile-amber-to text-canvas",
  rose: "from-tile-rose-from to-tile-rose-to text-canvas",
};

export function IconTile({ icon: Icon, tone, className }: { icon: ComponentType<{ className?: string }>; tone: TileTone; className?: string }) {
  return (
    <span className={cx("flex size-10 shrink-0 items-center justify-center rounded-2lg bg-linear-to-b", tones[tone], className)}>
      <Icon className="size-5" />
    </span>
  );
}
