import { useEffect, useState } from "react";
import * as Service from "@wails/go/app/Service";
import { AppProvider, useApp } from "./state";
import { Shell } from "./components/Shell";
import { Home } from "./screens/Home";
import { Diagnose } from "./screens/Diagnose";
import { Optimize } from "./screens/Optimize";
import { Flags } from "./screens/Flags";
import { Tests } from "./screens/Tests";
import { Settings } from "./screens/Settings";
import { Install } from "./screens/Install";
import type { InstallState } from "./screens/Install";

function Current() {
  const { screen } = useApp();
  switch (screen) {
    case "diagnose":
      return <Diagnose />;
    case "optimize":
      return <Optimize />;
    case "flags":
      return <Flags />;
    case "tests":
      return <Tests />;
    case "settings":
      return <Settings />;
    default:
      return <Home />;
  }
}

export default function App() {
  const [install, setInstall] = useState<InstallState | null>(null);
  const [skip, setSkip] = useState(false);
  useEffect(() => {
    Service.InstallState()
      .then((s) => setInstall(s as InstallState))
      .catch(() => setSkip(true));
  }, []);
  if (!install && !skip) return <div className="h-full bg-canvas" />;
  const needsInstall = install && !install.runningInstalled && (!install.installed || install.newer);
  if (needsInstall && !skip) return <Install state={install!} onSkip={() => setSkip(true)} />;
  return (
    <AppProvider>
      <Shell>
        <Current />
      </Shell>
    </AppProvider>
  );
}
