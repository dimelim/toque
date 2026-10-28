import { useEffect, useState } from "react";
import { api, onSession } from "@/lib/api";
import type { SessionInfo } from "@/lib/types";

export function useSession() {
  const [info, setInfo] = useState<SessionInfo | null>(null);
  useEffect(() => {
    let alive = true;
    const load = () =>
      api
        .session()
        .then((s) => alive && setInfo(s))
        .catch(() => undefined);
    load();
    const offSession = onSession((s) => alive && setInfo(s));
    return () => {
      alive = false;
      offSession();
    };
  }, []);
  return [info, setInfo] as const;
}
