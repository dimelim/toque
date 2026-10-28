import { useEffect, useState } from "react";
import { api, errorText } from "@/lib/api";
import { setLang, t } from "@/lib/i18n";
import { play } from "@/lib/sound";
import type { LanguageChoice, LanguageInfo } from "@/lib/types";
import { useApp } from "@/state";
import { Card } from "@/components/primitives";
import { SegmentedControl, SegmentedControlItem } from "@/components/ui/segmented-control";
import { IconGlobe } from "@/components/icons";

const names = { es: "Español", pt: "Português", en: "English" } as const;

export function LanguageCard() {
  const { toast, refresh } = useApp();
  const [info, setInfo] = useState<LanguageInfo | null>(null);
  useEffect(() => {
    api
      .language()
      .then(setInfo)
      .catch(() => undefined);
  }, []);
  const choose = async (choice: LanguageChoice) => {
    play("tap");
    try {
      const next = await api.setLanguage(choice);
      setInfo(next);
      setLang(next.active);
      refresh();
    } catch (e) {
      toast({ tone: "error", title: t("No pude cambiar el idioma"), detail: errorText(e) });
    }
  };
  return (
    <Card className="flex flex-col p-5">
      <p className="flex items-center gap-1.5 text-body-medium text-ink">
        <IconGlobe className="size-4 text-ink-3" />
        {t("Idioma")}
      </p>
      <p className="mt-1 text-body-regular text-ink-2">
        {info?.choice === "auto"
          ? t("Toque eligió {lang} por el idioma de tu Windows y el país desde el que te conectas.", { lang: names[info.detected] })
          : t("Si lo dejas en automático, Toque lo elige según tu Windows y tu país.")}
      </p>
      <div className="mt-auto pt-4">
        <SegmentedControl
          selectedKeys={info ? [info.choice] : []}
          onSelectionChange={(keys) => {
            const next = [...keys][0] as LanguageChoice | undefined;
            if (next && next !== info?.choice) choose(next);
          }}
          aria-label={t("Idioma")}
        >
          <SegmentedControlItem id="auto">{t("Automático")}</SegmentedControlItem>
          {(Object.keys(names) as (keyof typeof names)[]).map((l) => (
            <SegmentedControlItem key={l} id={l}>
              {names[l]}
            </SegmentedControlItem>
          ))}
        </SegmentedControl>
      </div>
    </Card>
  );
}
