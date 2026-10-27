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
