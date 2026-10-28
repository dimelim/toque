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
}

type windowsCopy struct {
	group, title, what string
}

var windowsText = map[string]windowsCopy{
	"privacy:telemetry":   {"datos", "Telemetría de Windows", "Deja el envío de diagnósticos en el mínimo obligatorio y desactiva el servicio que los sube, Experiencias del usuario conectado y telemetría."},
	"privacy:typing":      {"datos", "Datos de escritura y lápiz", "Windows deja de recopilar lo que escribes y dibujas para afinar sus sugerencias."},
	"privacy:feedback":    {"datos", "Preguntas de opinión", "Windows deja de pedirte que valores cómo funciona."},
	"privacy:ads":         {"anuncios", "Publicidad personalizada", "Apaga el identificador de publicidad y las experiencias personalizadas con tus datos de diagnóstico."},
	"privacy:suggestions": {"anuncios", "Sugerencias y apps automáticas", "Quita consejos y sugerencias de Inicio, Configuración y la pantalla de bloqueo, y evita que Windows instale por su cuenta apps que te sugiere."},
	"privacy:search":      {"anuncios", "Búsqueda web en Inicio", "La búsqueda del menú Inicio deja de mandar lo que escribes a Bing. Se nota al volver a iniciar sesión."},
}

func (s *Service) WindowsOptions() []WindowsOption {
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
		out = append(out, WindowsOption{ID: id, Group: c.group, Title: c.title, What: c.what, Applied: st.Applied, Managed: mine, Applicable: st.Applicable})
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
	if on {
		return tweaks.Apply(tweaks.Context{}, t)
	}
	return tweaks.Revert(tweaks.Context{}, t)
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
