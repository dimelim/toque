import { Button as AriaButton, ListBox, ListBoxItem, Popover, Select } from "react-aria-components";
import { num } from "@/lib/format";
import { PRIVATE, bestRegion, regionCity, regionPlace } from "@/lib/regions";
import type { Region } from "@/lib/types";
import { cx } from "@/utils/cx";
import { IconCheck, IconChevronDown, IconLock, IconPin } from "./icons";

const AUTO = "auto";

function Ping({ ms, size }: { ms: number | null | undefined; size: "lg" | "sm" }) {
  if (ms === null) return null;
  if (!ms) return <span className="shrink-0 text-caption-1-medium text-ink-4">sin medir</span>;
  return (
    <span className={cx("num shrink-0 leading-none font-semibold", size === "lg" ? "text-[26px] text-volt" : "text-[15px] text-ink")}>
      {num(ms, 0)}
      <span className="ml-1 font-sans text-[12px] font-medium text-ink-3">ms</span>
    </span>
  );
}

function Option({ id, title, place, ms }: { id: string; title: string; place: string; ms: number | null | undefined }) {
  return (
    <ListBoxItem
      id={id}
      textValue={title}
      className={({ isFocused }) => cx("flex cursor-default items-center gap-3 rounded-lg px-3 py-2 outline-none", isFocused && "bg-surface-2")}
    >
      {({ isSelected }) => (
        <>
          <IconCheck className={cx("size-4 shrink-0 text-volt", !isSelected && "invisible")} />
          <span className="min-w-0 flex-1">
            <span className="block truncate text-body-medium text-ink">{title}</span>
            <span className="block truncate text-caption-1-medium text-ink-3">{place}</span>
          </span>
          <Ping ms={ms} size="sm" />
        </>
      )}
    </ListBoxItem>
  );
}

export function RegionSelect({ regions, value, onChange, isDisabled }: { regions: Region[]; value: string; onChange: (key: string) => void; isDisabled?: boolean }) {
  const best = bestRegion(regions);
  const current = regions.find((r) => r.key === value);
  const isPrivate = value === PRIVATE;
  const Icon = isPrivate ? IconLock : IconPin;
  return (
    <Select aria-label="Región del servidor" value={value || AUTO} onChange={(k) => onChange(k === null || k === AUTO ? "" : String(k))} isDisabled={isDisabled}>
      <AriaButton className="flex w-full items-center gap-3 rounded-2lg bg-background-inner-default px-4 py-3 text-left shadow-card outline-none transition-colors duration-150 hover:bg-surface-2 focus-visible:ring-2 focus-visible:ring-border-focus-ring disabled:opacity-60">
        <span className="min-w-0 flex-1">
          <span className="flex items-center gap-1 text-caption-1-medium text-ink-3">
            <Icon className="size-3.5" />
            Región
          </span>
          <span className="mt-0.5 block truncate text-body-medium text-ink">{isPrivate ? "Servidor privado" : value ? (current?.city ?? regionCity(value)) : "Automática"}</span>
        </span>
        <Ping ms={isPrivate ? null : value ? current?.rttMs : best?.rttMs} size="lg" />
        <IconChevronDown className="size-4 shrink-0 text-ink-3" />
      </AriaButton>
      <Popover
        offset={6}
        className="w-[var(--trigger-width)] rounded-xl bg-surface-3 shadow-dropdown outline-none transition-[opacity,transform] duration-150 ease-out data-[entering]:-translate-y-1 data-[entering]:opacity-0 data-[exiting]:opacity-0"
      >
        <ListBox className="max-h-80 overflow-y-auto p-1.5 outline-none">
          <Option id={AUTO} title="Automática" place={best ? `Va al más cercano, ${best.city}` : "Va al servidor más cercano"} ms={best?.rttMs} />
          <Option id={PRIVATE} title="Servidor privado" place="Con su enlace de invitación" ms={null} />
          {regions.map((r) => (
            <Option key={r.key} id={r.key} title={r.city} place={regionPlace(r)} ms={r.rttMs} />
          ))}
        </ListBox>
        <p className="px-4 py-2.5 text-caption-1-medium text-ink-4 shadow-[inset_0_1px_0_var(--color-line)]">
          Roblox no tiene servidores en Perú ni en Chile. En Sudamérica solo está São Paulo. Ciudades según ipinfo.io.
        </p>
      </Popover>
    </Select>
  );
}
