import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Repo } from "@/lib/types";

export function useRepo() {
  const [repo, setRepo] = useState<Repo | null>(null);
  useEffect(() => {
    let alive = true;
    const load = () =>
      api
        .repo()
        .then((r) => alive && setRepo(r))
        .catch(() => {});
    load();
    const timer = setInterval(load, 30 * 60 * 1000);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, []);
  return repo;
}
