package app

import (
	"errors"

	"toque/internal/tweaks"
)

type WindowsOption struct {
	ID         string `json:"id"`
	Group      string `json:"group"`
	Title      string `json:"title"`
	What       string `json:"what"`
	Applied    bool   `json:"applied"`
	Managed    bool   `json:"managed"`
	Applicable bool   `json:"applicable"`
	Note       string `json:"note"`
}

type windowsCopy struct {
	group, title, what string
}

var windowsText = map[string]windowsCopy{
	"privacy:telemetry":   {"datos", "Datos de diagnóstico opcionales", "Apaga el envío de datos de diagnóstico opcionales. Windows sigue mandando el mínimo obligatorio, que usa para no ofrecerte actualizaciones con fallos conocidos en tu equipo. Mientras esté puesto, Configuración puede decir que tu organización administra este ajuste."},
	"privacy:typing":      {"datos", "Datos de escritura y lápiz", "Windows deja de recopilar lo que escribes y dibujas para afinar sus sugerencias."},
	"privacy:feedback":    {"datos", "Preguntas de opinión", "Windows deja de pedirte que valores cómo funciona."},
	"privacy:ads":         {"anuncios", "Publicidad personalizada", "Apaga el identificador de publicidad y las experiencias personalizadas con tus datos de diagnóstico."},
	"privacy:suggestions": {"anuncios", "Sugerencias y apps automáticas", "Quita consejos, recomendaciones y avisos de Inicio, Configuración, el Explorador y la pantalla de bloqueo, y evita que Windows instale por su cuenta apps que te sugiere."},
	"privacy:search":      {"anuncios", "Búsqueda web en Inicio", "La búsqueda del menú Inicio deja de mandar lo que escribes a Bing. Se nota al volver a iniciar sesión."},
	"privacy:widgets":     {"funciones", "Widgets y noticias", "Quita el panel de widgets, que mantiene procesos web abiertos en segundo plano."},
	"privacy:ai":          {"funciones", "Recall y Copilot antiguo", "Evita que Recall guarde capturas de lo que haces en los equipos que lo traen, y quita el Copilot lateral de Windows 11 23H2. El Copilot nuevo es una app normal que puedes desinstalar."},
}

func (s *Service) WindowsOptions() []WindowsOption {
	return windowsOptions()
}

func windowsOptions() []WindowsOption {
	managed, _ := tweaks.Entries()
	out := []WindowsOption{}
	for _, id := range tweaks.PrivacyIDs() {
		t, err := tweaks.Get(id)
		if err != nil {
			continue
		}
		st := t.State(tweaks.Context{})
		_, mine := managed[id]
		c := windowsText[id]
		out = append(out, WindowsOption{ID: id, Group: c.group, Title: c.title, What: c.what, Applied: st.Applied, Managed: mine, Applicable: st.Applicable, Note: st.Detail})
	}
	return out
}

func setWindowsOption(id string, on bool) error {
	if _, ok := windowsText[id]; !ok {
		return errors.New("no conozco esa opción de Windows")
	}
	t, err := tweaks.Get(id)
	if err != nil {
		return err
	}
	if !on {
		return tweaks.Revert(tweaks.Context{}, t)
	}
	if managed, _ := tweaks.Entries(); managed != nil {
		if _, mine := managed[id]; mine {
			return t.Apply(tweaks.Context{})
		}
	}
	return tweaks.Apply(tweaks.Context{}, t)
}

func (s *Service) SetWindowsOption(id string, on bool) ([]WindowsOption, error) {
	if err := s.idle(); err != nil {
		return s.WindowsOptions(), err
	}
	err := setWindowsOption(id, on)
	return s.WindowsOptions(), err
}

func (s *Service) SetAllWindowsOptions(on bool) ([]WindowsOption, error) {
	if err := s.idle(); err != nil {
		return s.WindowsOptions(), err
	}
	var errs []error
	for _, o := range s.WindowsOptions() {
		switch {
		case on && o.Applicable && !o.Applied:
			errs = append(errs, setWindowsOption(o.ID, true))
		case !on && o.Managed:
			errs = append(errs, setWindowsOption(o.ID, false))
		}
	}
	return s.WindowsOptions(), errors.Join(errs...)
}

func pendingPrivacy() (pending, unknown int) {
	for _, o := range windowsOptions() {
		switch {
		case !o.Applicable && !o.Applied:
			unknown++
		case !o.Applied:
			pending++
		}
	}
	return pending, unknown
}
