import { t } from "@/lib/i18n";
import { useStored } from "@/state";
import { SegmentedControl, SegmentedControlItem } from "@/components/ui/segmented-control";
import { Optimize } from "./Optimize";
import { Flags } from "./Flags";
import { WindowsTab } from "./optimize/WindowsTab";

const tabs = () => [
  { id: "tested", label: t("Ajustes probados") },
  { id: "windows", label: "Windows" },
  { id: "flags", label: t("Flags de Roblox") },
];

export function OptimizeArea() {
  const [stored, setTab] = useStored("toque.optimizeTab");
  const items = tabs();
  const tab = items.some((item) => item.id === stored) ? stored : "tested";
  return (
    <div>
      <SegmentedControl
        className="mb-6"
        selectedKeys={[tab]}
        onSelectionChange={(keys) => {
          const id = [...keys][0];
          if (id !== undefined) setTab(String(id));
        }}
        aria-label={t("Secciones de Optimizar")}
      >
        {items.map((item) => (
          <SegmentedControlItem key={item.id} id={item.id}>
            {item.label}
          </SegmentedControlItem>
        ))}
      </SegmentedControl>
      {tab === "windows" ? <WindowsTab /> : tab === "flags" ? <Flags /> : <Optimize />}
    </div>
  );
}
