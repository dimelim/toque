import { motion } from "motion/react";

export function Wordmark({ className = "" }: { className?: string }) {
  return (
    <svg viewBox="-8 -8 808 272" className={className} role="img" aria-label="Toque" fill="currentColor">
      <path d="M24 0H60V196H24Z" />
      <path d="M0 44H96V76H0Z" />
      <path fillRule="evenodd" d="M112 120A76 76 0 1 1 264 120A76 76 0 1 1 112 120ZM148 120A40 40 0 1 0 228 120A40 40 0 1 0 148 120Z" />
      <path fillRule="evenodd" d="M284 120A76 76 0 1 1 436 120A76 76 0 1 1 284 120ZM320 120A40 40 0 1 0 400 120A40 40 0 1 0 320 120Z" />
      <path d="M400 120H436V256L400 227.2Z" />
      <path d="M464 44V120A76 76 0 0 0 540 196H616V44H580V120A40 40 0 0 1 500 120V44Z" />
      <path fillRule="evenodd" d="M788.33 135.84A76 76 0 1 0 772.22 168.85L744.64 145.71A40 40 0 0 1 677.27 135.84H788.33ZM677.27 104.16A40 40 0 0 1 750.73 104.16Z" />
    </svg>
  );
}

export function Symbol({ className = "" }: { className?: string }) {
  return (
    <svg viewBox="28 36 168 228" className={className} role="img" aria-label="Toque" fill="currentColor">
      <path fillRule="evenodd" d="M36 120A76 76 0 1 1 188 120A76 76 0 1 1 36 120ZM72 120A40 40 0 1 0 152 120A40 40 0 1 0 72 120Z" />
      <path d="M152 120H188V256L152 227.2Z" />
    </svg>
  );
}

export function AppIcon({ className = "" }: { className?: string }) {
  return (
    <span className={`grid place-items-center rounded-[22%] bg-volt text-volt-ink ${className}`}>
      <Symbol className="h-[62%]" />
    </span>
  );
}

const ring = "M36 120A76 76 0 1 1 188 120A76 76 0 1 1 36 120ZM72 120A40 40 0 1 0 152 120A40 40 0 1 0 72 120Z";
const tail = "M152 120H188V256L152 227.2Z";
const drawEase = [0.65, 0, 0.35, 1] as const;

export function SymbolReveal({ className = "" }: { className?: string }) {
  return (
    <svg viewBox="26 34 172 232" className={className} role="img" aria-label="Toque">
      <motion.path
        d={ring}
        fillRule="evenodd"
        className="fill-volt stroke-volt"
        strokeWidth="1.5"
        initial={{ pathLength: 0, fillOpacity: 0 }}
        animate={{ pathLength: 1, fillOpacity: 1 }}
        transition={{ pathLength: { duration: 1.2, ease: drawEase }, fillOpacity: { duration: 0.6, delay: 1.05, ease: "easeOut" } }}
      />
      <motion.path
        d={tail}
        className="fill-volt stroke-volt"
        strokeWidth="1.5"
        initial={{ pathLength: 0, fillOpacity: 0 }}
        animate={{ pathLength: 1, fillOpacity: 1 }}
        transition={{ pathLength: { duration: 0.7, delay: 0.75, ease: drawEase }, fillOpacity: { duration: 0.6, delay: 1.25, ease: "easeOut" } }}
      />
    </svg>
  );
}
