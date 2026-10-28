import { useStored } from "@/state";
import { SegmentedControl, SegmentedControlItem } from "@/components/ui/segmented-control";
import { Optimize } from "./Optimize";
import { Flags } from "./Flags";
import { WindowsTab } from "./optimize/WindowsTab";

const tabs = [
  { id: "tested", label: "Ajustes probados" },
  { id: "windows", label: "Windows" },
  { id: "flags", label: "Flags de Roblox" },
];

export function OptimizeArea() {
  const [stored, setTab] = useStored("toque.optimizeTab");
  const tab = tabs.some((t) => t.id === stored) ? stored : "tested";
  return (
    <div>
      <SegmentedControl
        className="mb-6"
        selectedKeys={[tab]}
        onSelectionChange={(keys) => {
          const id = [...keys][0];
          if (id !== undefined) setTab(String(id));
        }}
        aria-label="Secciones de Optimizar"
      >
        {tabs.map((t) => (
          <SegmentedControlItem key={t.id} id={t.id}>
            {t.label}
          </SegmentedControlItem>
        ))}
      </SegmentedControl>
      {tab === "windows" ? <WindowsTab /> : tab === "flags" ? <Flags /> : <Optimize />}
    </div>
  );
}
