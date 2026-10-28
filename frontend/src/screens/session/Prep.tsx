import { motion } from "motion/react";
import { cx } from "@/utils/cx";
import type { PrepCheck } from "@/lib/types";
import { Card } from "@/components/primitives";
import { Button } from "@/components/ui/button";

function Check({ done }: { done: boolean }) {
  return (
    <svg viewBox="0 0 28 28" className="size-7 shrink-0" aria-hidden>
      <circle cx="14" cy="14" r="12" fill="none" strokeWidth="2" className={cx("transition-[stroke] duration-300", done ? "stroke-volt" : "stroke-line-strong")} />
      <motion.path
        d="M8.5 14.5 12.5 18.5 19.5 10"
        fill="none"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="stroke-volt"
        initial={false}
        animate={{ pathLength: done ? 1 : 0, opacity: done ? 1 : 0 }}
        transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1] }}
      />
    </svg>
  );
}

export function Prep({ checks, disabled, onAct }: { checks: PrepCheck[]; disabled: boolean; onAct: (c: PrepCheck) => void }) {
  return (
    <div className={cx("grid gap-3", checks.length > 1 ? "grid-cols-2" : "grid-cols-1")}>
      {checks.map((c) => (
        <Card key={c.id} className="flex items-start gap-4 p-5">
          <Check done={c.done} />
          <div className="flex min-w-0 flex-1 flex-col">
            <h3 className="text-[15px] font-semibold text-ink">{c.title}</h3>
            <p className="mt-1 text-body-regular text-ink-2">{c.detail}</p>
            {c.blocked ? <p className="mt-2 text-caption-1-medium text-warn">{c.blocked}</p> : null}
            {c.action ? (
              <div className="mt-4">
                <Button size="small" variant="secondary" disabled={disabled || !!c.blocked} onClick={() => onAct(c)}>
                  {c.action}
                </Button>
              </div>
            ) : null}
          </div>
        </Card>
      ))}
    </div>
  );
}
