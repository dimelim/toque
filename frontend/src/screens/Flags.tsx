import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Dialog, Heading, Input, Modal, ModalOverlay, Slider, SliderThumb, SliderTrack, TextArea, TextField } from "react-aria-components";
import { cx } from "@/utils/cx";
import { api, errorText } from "@/lib/api";
import { num } from "@/lib/format";
import { play } from "@/lib/sound";
import { FLAG_NAME, flagSections, flagType, flagTypeLabel, groupState, parseFlagsJSON, validValue } from "@/lib/flags";
import type { FlagGroup, FlagType, FlagValues } from "@/lib/flags";
import type { FlagEntry } from "@/lib/types";
import { useApp } from "@/state";
import { Card, Empty, Eyebrow, PageHeader, Pill } from "@/components/primitives";
import { Button } from "@/components/ui/button";
import { SegmentedControl, SegmentedControlItem } from "@/components/ui/segmented-control";
import { Switch } from "@/components/ui/switch";
import { IconCopy, IconImport, IconPlus, IconSearch, IconTrash } from "@/components/icons";

type Update = (set: FlagValues, clear: string[], done?: string) => Promise<boolean>;

function targetOf(group: FlagGroup, id: string): FlagValues {
  if (group.control === "switch") return id === "on" ? group.on : {};
  if (group.control === "level") return id === "0" ? {} : { [group.keys[0]]: id };
  return group.options.find((o) => o.id === id)?.values ?? {};
}

function LevelSlider({ group, value, disabled, onCommit }: { group: FlagGroup & { control: "level" }; value: number; disabled: boolean; onCommit: (v: number) => void }) {
  const [local, setLocal] = useState(value);
  useEffect(() => setLocal(value), [value]);
  return (
    <Slider
      className="mt-4"
      minValue={0}
      maxValue={group.max}
      step={1}
      value={local}
      onChange={(v) => setLocal(v as number)}
      onChangeEnd={(v) => onCommit(v as number)}
      isDisabled={disabled}
      aria-label={group.title}
    >
      <div className="flex items-baseline justify-between">
        <span className="text-caption-1-medium text-ink-3">{local === 0 ? "La decide Roblox" : `Nivel ${local} de ${group.max}`}</span>
        <span className={cx("num text-[22px] leading-none font-semibold", local === 0 ? "text-ink-2" : "text-volt")}>{local === 0 ? "Auto" : local}</span>
      </div>
      <SliderTrack className="relative mt-2 h-6 w-full cursor-pointer data-[disabled]:cursor-not-allowed">
        {({ state }) => (
          <>
            <div className="absolute top-1/2 h-1 w-full -translate-y-1/2 rounded-full bg-surface-3" />
            <div className="absolute top-1/2 h-1 -translate-y-1/2 rounded-full bg-volt" style={{ width: `${state.getThumbPercent(0) * 100}%` }} />
            <SliderThumb className="top-1/2 size-[18px] -translate-y-1/2 rounded-full bg-ink shadow-[0_1px_4px_rgb(0_0_0/0.45)] outline-none transition-[scale] duration-150 data-[dragging]:scale-110 data-[focus-visible]:ring-2 data-[focus-visible]:ring-border-focus-ring" />
          </>
        )}
      </SliderTrack>
    </Slider>
  );
}

function GroupCard({ group, values, disabled, update }: { group: FlagGroup; values: FlagValues; disabled: boolean; update: Update }) {
  const state = groupState(group, values);
  const choose = (id: string) => {
    const target = targetOf(group, id);
    update(target, group.keys.filter((k) => !(k in target)));
  };
  return (
    <Card className="flex min-w-0 flex-col p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="text-[16px] leading-snug font-semibold text-ink">{group.title}</h3>
            {state.custom ? <Pill>Valor propio</Pill> : state.active ? <Pill tone="volt">Activa</Pill> : null}
          </div>
          <p className="mt-1 text-body-regular text-ink-2">{group.text}</p>
        </div>
        {group.control === "switch" ? (
          <Switch size="md" className="mt-0.5 shrink-0" isSelected={state.selected === "on"} isDisabled={disabled} onChange={(on) => choose(on ? "on" : "off")} aria-label={group.title} />
        ) : null}
      </div>
      {group.control === "choice" ? (
        <SegmentedControl
          className="mt-4 self-start"
          selectedKeys={state.selected ? [state.selected] : []}
          onSelectionChange={(keys) => {
            const id = [...keys][0];
            if (id !== undefined && id !== state.selected) choose(String(id));
          }}
          isDisabled={disabled}
          aria-label={group.title}
        >
          {group.options.map((o) => (
            <SegmentedControlItem key={o.id} id={o.id}>
              {o.label}
            </SegmentedControlItem>
          ))}
        </SegmentedControl>
      ) : null}
      {group.control === "level" ? <LevelSlider group={group} value={Number(state.selected ?? 0)} disabled={disabled} onCommit={(v) => choose(String(v))} /> : null}
      <div className="mt-auto pt-4">
        {state.custom ? <p className="mb-1.5 text-caption-1-medium text-ink-3">Valor actual: {state.custom}.</p> : null}
        {group.note ? <p className="text-caption-1-medium text-ink-3">{group.note}</p> : null}
        <p className="mt-2 truncate font-mono text-[11px] text-ink-4" title={group.keys.join("\n")}>
          {group.keys.join(" · ")}
        </p>
      </div>
    </Card>
  );
}

const valueHint: Record<FlagType, string> = {
  bool: "True o False",
  int: "Un número entero",
  log: "Un número entero",
  string: "Texto",
  unknown: "Valor",
};

function ValueCell({ flag, disabled, update }: { flag: FlagEntry; disabled: boolean; update: Update }) {
  const type = flagType(flag.name);
  const [draft, setDraft] = useState(flag.value);
  useEffect(() => setDraft(flag.value), [flag.value]);
  if (type === "bool" && validValue(type, flag.value)) {
    return (
      <div className="flex h-8 w-[30%] shrink-0 items-center justify-end gap-2">
        <span className="font-mono text-[13px] text-ink-2">{/^true$/i.test(flag.value) ? "True" : "False"}</span>
        <Switch size="sm" isSelected={/^true$/i.test(flag.value)} isDisabled={disabled} onChange={(on) => update({ [flag.name]: on ? "True" : "False" }, [])} aria-label={`Valor de ${flag.name}`} />
      </div>
    );
  }
  const valid = validValue(type, draft.trim());
  const commit = () => {
    if (!valid) {
      setDraft(flag.value);
      return;
    }
    if (draft.trim() !== flag.value) update({ [flag.name]: draft.trim() }, []);
  };
  return (
    <TextField aria-label={`Valor de ${flag.name}`} value={draft} onChange={setDraft} isDisabled={disabled} className="w-[30%] shrink-0">
      <Input
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === "Enter") e.currentTarget.blur();
          if (e.key === "Escape") setDraft(flag.value);
        }}
        title={valid ? flag.value : valueHint[type]}
        className={cx(
          "h-8 w-full rounded-md bg-transparent px-2 text-right font-mono text-[13px] text-ink-2 outline-none transition-colors duration-150 hover:bg-background-inner-default focus:bg-background-inner-default focus:text-ink focus:ring-2",
          valid ? "focus:ring-border-focus-ring" : "text-bad ring-2 ring-bad",
        )}
      />
    </TextField>
  );
}

function FlagEditor({ flags, disabled, update }: { flags: FlagEntry[]; disabled: boolean; update: Update }) {
  const [name, setName] = useState("");
  const [value, setValue] = useState("");
  const [filter, setFilter] = useState("");
  const [confirm, setConfirm] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);
  useEffect(() => () => clearTimeout(timer.current), []);
  const type = flagType(name.trim());
  const validName = FLAG_NAME.test(name.trim());
  const validVal = validValue(type, value.trim());
  const valid = validName && validVal;
  const shown = filter ? flags.filter((f) => f.name.toLowerCase().includes(filter.toLowerCase())) : flags;
  const add = async () => {
    if (!valid) return;
    const v = type === "bool" ? (/^true$/i.test(value.trim()) ? "True" : "False") : value.trim();
    if (await update({ [name.trim()]: v }, [], `${name.trim()} guardada`)) {
      setName("");
      setValue("");
    }
  };
  const clearAll = () => {
    if (!confirm) {
      setConfirm(true);
      timer.current = setTimeout(() => setConfirm(false), 3000);
      return;
    }
    clearTimeout(timer.current);
    setConfirm(false);
    update({}, flags.map((f) => f.name), "Quitadas todas las flags");
  };
  const field = "h-10 w-full rounded-lg bg-background-inner-default px-3 font-mono text-[13px] text-ink shadow-card outline-none placeholder:font-sans placeholder:text-ink-4 focus:ring-2 focus:ring-border-focus-ring";
  return (
    <div className="mt-10">
      <div className="mb-3 flex items-baseline justify-between gap-4">
        <h2 className="text-[15px] font-semibold text-ink">Editor de flags</h2>
        <p className="text-caption-1-medium text-ink-3">Cualquier flag de cualquier tipo, como en Bloxstrap o Fishstrap</p>
      </div>
      <Card className="p-3">
        <form
          className="grid grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_auto] gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            add();
          }}
        >
          <TextField aria-label="Nombre de la flag" value={name} onChange={setName} isDisabled={disabled}>
            <Input placeholder="Nombre, por ejemplo FFlagDebugDisplayFPS" className={field} />
          </TextField>
          <TextField aria-label="Valor" value={value} onChange={setValue} isDisabled={disabled}>
            <Input placeholder={name.trim() ? valueHint[type] : "Valor, por ejemplo True"} className={cx(field, value.trim() && !validVal && "ring-2 ring-bad")} />
          </TextField>
          <Button type="submit" leadingIcon={IconPlus} disabled={disabled || !valid}>
            Añadir
          </Button>
        </form>
        {validName && type === "unknown" ? (
          <p className="mt-2 px-1 text-caption-1-medium text-ink-3">Las flags de Roblox empiezan por FFlag, FInt, FString o FLog, con D o S delante o sin ella.</p>
        ) : null}
        {validName && value.trim() && !validVal ? <p className="mt-2 px-1 text-caption-1-medium text-bad">Esta flag espera {valueHint[type].toLowerCase()}.</p> : null}
        {flags.length > 8 ? (
          <div className="mt-3 flex items-center gap-2">
            <TextField aria-label="Buscar flags" value={filter} onChange={setFilter} className="relative min-w-0 flex-1">
              <IconSearch className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-ink-3" aria-hidden />
              <Input placeholder={`Buscar entre ${num(flags.length)} flags`} className={cx(field, "pl-9 font-sans")} />
            </TextField>
            <Button variant={confirm ? "danger" : "ghost"} leadingIcon={IconTrash} disabled={disabled} onClick={clearAll}>
              {confirm ? "Pulsa otra vez para quitarlas" : "Quitar todas"}
            </Button>
          </div>
        ) : null}
        {flags.length === 0 ? (
          <p className="px-1 pt-3 text-body-regular text-ink-3">Todavía no tienes flags. Añade una arriba o importa una lista en JSON.</p>
        ) : (
          <ul className="mt-2 max-h-[420px] overflow-y-auto">
            {shown.map((f) => (
              <li key={f.name} className="flex items-center gap-3 rounded-lg px-2 py-2 transition-colors duration-150 hover:bg-surface-2">
                <span className="w-[76px] shrink-0 text-caption-1-medium text-ink-3">{flagTypeLabel[flagType(f.name)]}</span>
                <span className="min-w-0 flex-1 truncate font-mono text-[13px] text-ink" title={f.name}>
                  {f.name}
                </span>
                <ValueCell flag={f} disabled={disabled} update={update} />
                <Button variant="ghost" size="xs" iconOnly leadingIcon={IconTrash} aria-label={`Quitar ${f.name}`} disabled={disabled} onClick={() => update({}, [f.name])} />
              </li>
            ))}
            {shown.length === 0 ? <li className="px-2 py-3 text-body-regular text-ink-3">Ninguna contiene «{filter}».</li> : null}
          </ul>
        )}
      </Card>
    </div>
  );
}

function ImportDialog({ open, disabled, onClose, update }: { open: boolean; disabled: boolean; onClose: () => void; update: Update }) {
  const [text, setText] = useState("");
  const file = useRef<HTMLInputElement>(null);
  const parsed = useMemo(() => parseFlagsJSON(text), [text]);
  const readFile = async (picked: File | undefined) => {
    if (picked) setText(await picked.text());
    if (file.current) file.current.value = "";
  };
  const close = () => {
    setText("");
    onClose();
  };
  const submit = async () => {
    if (parsed.state !== "ok" || parsed.count === 0) return;
    if (await update(parsed.values, [], `Importadas ${num(parsed.count)} flags`)) close();
  };
  return (
    <ModalOverlay
      isOpen={open}
      onOpenChange={(v) => !v && close()}
      isDismissable
      className="fixed inset-0 z-50 grid place-items-center bg-black/55 p-6 backdrop-blur-[2px] transition-opacity duration-200 data-[entering]:opacity-0 data-[exiting]:opacity-0"
    >
      <Modal className="w-full max-w-[640px] rounded-2xl bg-surface-3 shadow-dropdown outline-none transition-[scale,opacity] duration-200 ease-out data-[entering]:scale-[0.97] data-[entering]:opacity-0">
        <Dialog className="flex flex-col gap-4 p-6 outline-none">
          <div>
            <Heading slot="title" className="text-[19px] font-semibold text-ink">
              Importar flags
            </Heading>
            <p className="mt-1 text-body-regular text-ink-2">Pega un JSON como los de Bloxstrap o Fishstrap, o abre el archivo. Se suman a las flags que ya tienes.</p>
          </div>
          <TextField aria-label="JSON de flags" value={text} onChange={setText} autoFocus>
            <TextArea
              rows={12}
              spellCheck={false}
              placeholder={'{\n  "DFFlagDebugPauseVoxelizer": "True",\n  "FFlagDebugSkyGray": "True"\n}'}
              className="w-full resize-none rounded-xl bg-background-inner-default p-3 font-mono text-[12.5px] leading-relaxed text-ink shadow-card outline-none placeholder:text-ink-4 focus:ring-2 focus:ring-border-focus-ring"
            />
          </TextField>
          <div className="min-h-6">
            {parsed.state === "invalid" ? <p className="text-body-medium text-bad">Eso no es un JSON de flags. Tiene que ser un objeto con nombre y valor.</p> : null}
            {parsed.state === "ok" ? (
              <div className="flex flex-wrap items-center gap-1.5">
                <Pill tone="volt">{num(parsed.count)} flags</Pill>
                {parsed.mistyped > 0 ? <Pill>{num(parsed.mistyped)} con un valor de otro tipo</Pill> : null}
                {parsed.skipped > 0 ? <Pill>{num(parsed.skipped)} no son flags y se saltan</Pill> : null}
              </div>
            ) : null}
          </div>
          <div className="flex items-center gap-2">
            <input ref={file} type="file" accept=".json,.txt,application/json,text/plain" className="hidden" onChange={(e) => readFile(e.target.files?.[0])} />
            <Button variant="ghost" leadingIcon={IconImport} onClick={() => file.current?.click()}>
              Abrir archivo
            </Button>
            <span className="flex-1" />
            <Button variant="ghost" onClick={close}>
              Cancelar
            </Button>
            <Button leadingIcon={IconImport} disabled={disabled || parsed.state !== "ok" || parsed.count === 0} onClick={submit}>
              {parsed.state === "ok" && parsed.count > 0 ? `Importar ${num(parsed.count)}` : "Importar"}
            </Button>
          </div>
        </Dialog>
      </Modal>
    </ModalOverlay>
  );
}

export function Flags() {
  const { task, toast, overview } = useApp();
  const [flags, setFlags] = useState<FlagEntry[] | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [importing, setImporting] = useState(false);
  const load = useCallback(async () => {
    try {
      setFlags(await api.flags());
      setError("");
    } catch (e) {
      setError(errorText(e));
    }
  }, []);
  useEffect(() => {
    load();
  }, [load]);
  const values = useMemo<FlagValues>(() => Object.fromEntries((flags ?? []).map((f) => [f.name, f.value])), [flags]);
  const disabled = busy || !!task || !flags;
  const update: Update = async (set, clear, done) => {
    setBusy(true);
    try {
      setFlags(await api.updateFlags(set, clear));
      play(Object.keys(set).length ? "on" : "off");
      if (done) toast({ tone: "ok", title: done });
      return true;
    } catch (e) {
      play("error");
      toast({ tone: "error", title: "No pude cambiar las flags", detail: errorText(e) });
      return false;
    } finally {
      setBusy(false);
    }
  };
  const copy = async () => {
    const json = JSON.stringify(values, null, 2);
    if (await api.copyText(json)) toast({ tone: "ok", title: `Copiadas ${num(flags?.length ?? 0)} flags en JSON` });
  };
  const active = flagSections.flatMap((s) => s.groups).filter((g) => groupState(g, values).active).length;

  if (error && !flags) {
    return (
      <div>
        <PageHeader eyebrow="Optimizar" title="Ajustes internos de Roblox" />
        <Empty title="No puedo leer las flags" action={<Button onClick={load}>Reintentar</Button>}>
          {error}
        </Empty>
      </div>
    );
  }

  return (
    <div>
      <PageHeader eyebrow="Optimizar" title="Ajustes internos de Roblox">
        <Button variant="secondary" leadingIcon={IconCopy} disabled={!flags?.length} onClick={copy}>
          Copiar JSON
        </Button>
        <Button leadingIcon={IconImport} disabled={disabled} onClick={() => setImporting(true)}>
          Importar JSON
        </Button>
      </PageHeader>
      <Card className="mb-8 flex items-center gap-6 p-5">
        <div className="flex-1">
          <Eyebrow>Cómo funciona</Eyebrow>
          <p className="mt-2 max-w-[78ch] text-body-regular text-ink-2">
            Arriba tienes los ajustes más usados con controles fáciles. Abajo está el editor con todas tus flags, donde puedes añadir cualquiera o importar una lista en JSON, como en Bloxstrap o Fishstrap.
          </p>
          <p className="mt-2 text-caption-1-medium text-ink-3">
            {overview?.roblox.running ? "Roblox está abierto. Los cambios se verán cuando lo vuelvas a abrir." : "Los cambios se aplican la próxima vez que abras Roblox."}
          </p>
        </div>
        <div className="flex gap-6 text-right">
          {[
            { n: active, label: "ajustes activos", tone: "text-volt" },
            { n: flags?.length ?? 0, label: "flags en total", tone: "text-ink-2" },
          ].map((c) => (
            <div key={c.label}>
              <p className={cx("num text-[34px] leading-none font-semibold", c.tone)}>{num(c.n)}</p>
              <p className="mt-1 text-caption-1-medium whitespace-nowrap text-ink-3">{c.label}</p>
            </div>
          ))}
        </div>
      </Card>
      {flagSections.map((s) => (
        <div key={s.id} className="mt-8 first-of-type:mt-0">
          <div className="mb-3 flex items-baseline justify-between gap-4">
            <h2 className="text-[15px] font-semibold text-ink">{s.title}</h2>
            <p className="text-caption-1-medium text-ink-3">{s.hint}</p>
          </div>
          <div className="grid grid-cols-2 gap-3">
            {s.groups.map((g) => (
              <GroupCard key={g.id} group={g} values={values} disabled={disabled} update={update} />
            ))}
          </div>
        </div>
      ))}
      <FlagEditor flags={flags ?? []} disabled={disabled} update={update} />
      <ImportDialog open={importing} disabled={disabled} onClose={() => setImporting(false)} update={update} />
    </div>
  );
}
