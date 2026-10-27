import type { HTMLAttributes } from "react";
import { cx, sortCx } from "@/utils/cx";

type StatusDotColor = "green" | "yellow" | "gray";

const styles = sortCx({
  halo: {
    green: "bg-status-dot-green-halo",
    yellow: "bg-status-dot-yellow-halo",
    gray: "bg-status-dot-gray-halo",
  },
  dot: {
    green: "bg-volt",
    yellow: "bg-warn",
    gray: "bg-ink-3",
  },
});

export function StatusDot({ color = "green", className, ...props }: HTMLAttributes<HTMLSpanElement> & { color?: StatusDotColor }) {
  return (
    <span aria-hidden className={cx("inline-flex size-3 shrink-0 items-center justify-center rounded-full", styles.halo[color], className)} {...props}>
      <span className={cx("size-1.5 rounded-full", styles.dot[color])} />
    </span>
  );
}
