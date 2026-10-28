import { useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { cx } from "@/utils/cx";
import type { SessionInfo, SessionState } from "@/lib/types";
import { Card } from "@/components/primitives";
import { IconRocket } from "@/components/icons";

const copy: Record<SessionState, { title: string; text: string }> = {
  off: { title: "Modo partida apagado", text: "Enciéndelo y Toque cuidará tu partida cada vez que abras Roblox." },
  waiting: { title: "Listo para tu próxima partida", text: "Se activa solo cuando abras Roblox." },
  active: { title: "Cuidando tu partida", text: "" },
  paused: { title: "En pausa", text: "En pausa mientras Toque hace una prueba." },
};

const iconTone: Record<SessionState, string> = {
  off: "text-ink-4",
  waiting: "text-ink-2",
  active: "text-volt",
  paused: "text-warn",
};

const layer = { initial: { opacity: 0 }, animate: { opacity: 1 }, exit: { opacity: 0 }, transition: { duration: 0.5, ease: [0.22, 1, 0.36, 1] as const } };

function useNow(active: boolean) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!active) return;
    setNow(Date.now());
    const t = setInterval(() => setNow(Date.now()), 15000);
    return () => clearInterval(t);
  }, [active]);
  return now;
}

function elapsed(since: string, now: number) {
  const start = new Date(since).getTime();
  if (Number.isNaN(start)) return "";
  const min = Math.max(0, Math.floor((now - start) / 60000));
  if (min < 1) return "Acabas de empezar";
  if (min < 60) return `Llevas ${min} min de partida`;
  return `Llevas ${Math.floor(min / 60)} h ${String(min % 60).padStart(2, "0")} min de partida`;
}

function Engine({ state }: { state: SessionState }) {
  return (
    <div className="relative size-[200px] shrink-0">
      <svg viewBox="0 0 200 200" className="absolute inset-0 size-full" aria-hidden>
        <circle cx="100" cy="100" r="94" className="fill-none stroke-line-strong" strokeWidth="2" strokeDasharray="1 9" strokeLinecap="round" />
        <circle cx="100" cy="100" r="78" className="fill-none stroke-surface-3" strokeWidth="3" />
        <AnimatePresence initial={false}>
          {state === "active" ? (
            <motion.g key="active" {...layer}>
              <g className="session-layer session-spin">
                <circle cx="100" cy="100" r="78" className="fill-none stroke-volt" strokeWidth="3.5" strokeLinecap="round" strokeDasharray="140 350" />
              </g>
              <g className="session-layer session-spin-slow">
                <circle cx="100" cy="100" r="62" className="fill-none stroke-volt" strokeOpacity="0.45" strokeWidth="2" strokeLinecap="round" strokeDasharray="70 320" />
              </g>
            </motion.g>
          ) : null}
          {state === "waiting" ? (
            <motion.g key="waiting" {...layer}>
              <g className="session-layer session-breathe">
                <circle cx="100" cy="100" r="78" className="fill-none stroke-volt" strokeWidth="2" />
              </g>
            </motion.g>
          ) : null}
          {state === "paused" ? (
            <motion.g key="paused" {...layer}>
              <circle cx="100" cy="100" r="78" className="fill-none stroke-warn" strokeWidth="2.5" strokeLinecap="round" strokeDasharray="6 12" />
            </motion.g>
          ) : null}
        </AnimatePresence>
      </svg>
      <div className="absolute inset-0 grid place-items-center">
        <motion.span
          animate={{ scale: state === "active" ? 1.08 : state === "off" ? 0.92 : 1 }}
          transition={{ type: "spring", duration: 0.5, bounce: 0.25 }}
          className={cx("transition-colors duration-500", iconTone[state])}
        >
          <IconRocket className="size-11" />
        </motion.span>
      </div>
    </div>
  );
}

export function Hero({ info }: { info: SessionInfo | null }) {
  const state: SessionState = info?.state ?? "off";
  const now = useNow(state === "active");
  const text = state === "active" && info ? elapsed(info.since, now) : copy[state].text;
  return (
    <Card className="mb-8 flex items-center gap-8 overflow-hidden p-6">
      <Engine state={state} />
      <div className="min-w-0 flex-1">
        <AnimatePresence mode="wait" initial={false}>
          <motion.div key={state} initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -4 }} transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}>
            <h2 className={cx("text-[26px] leading-[1.15] font-semibold tracking-[-0.02em]", state === "active" ? "text-volt" : "text-ink")}>{copy[state].title}</h2>
            <p className="num mt-2 text-body-regular text-ink-2">{text}</p>
          </motion.div>
        </AnimatePresence>
        <p className="mt-5 max-w-[60ch] text-caption-1-medium text-ink-3">
          Toque solo cambia lo que puede deshacer y te enseña qué te frena. Lo que no está medido lo marca como sin probar.
        </p>
      </div>
    </Card>
  );
}
