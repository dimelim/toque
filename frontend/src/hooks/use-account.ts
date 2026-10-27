import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Account } from "@/lib/types";

export function useAccount() {
  const [account, setAccount] = useState<Account | null>(null);
  useEffect(() => {
    let alive = true;
    const load = () =>
      api
        .account()
        .then((a) => alive && setAccount(a))
        .catch(() => alive && setAccount(null));
    load();
    const timer = setInterval(load, 5 * 60 * 1000);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, []);
  return account;
}
