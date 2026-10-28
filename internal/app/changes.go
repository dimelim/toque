package app

import (
	"fmt"
	"sort"
	"strings"

	"toque/internal/catalog"
	"toque/internal/tweaks"
)

type ActiveChange struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func changeTitle(id string) string {
	if c, ok := windowsText[id]; ok {
		return c.title
	}
	switch {
	case id == "windowed-games":
		return "Juegos en ventana"
	case id == "windows:gamemode":
		return "Modo de juego de Windows"
	case strings.HasPrefix(id, "setting:FramerateCap="):
		return "Límite de FPS a " + strings.TrimPrefix(id, "setting:FramerateCap=")
	}
	return catalog.Title(id)
}

func (s *Service) ActiveChanges() []ActiveChange {
	entries, _ := tweaks.Entries()
	out := []ActiveChange{}
	flags := 0
	for id := range entries {
		if _, known := catalog.Find(id); strings.HasPrefix(id, "flag:") && !known {
			flags++
			continue
		}
		out = append(out, ActiveChange{ID: id, Title: changeTitle(id)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Title < out[j].Title })
	switch {
	case flags == 1:
		out = append(out, ActiveChange{ID: "flags", Title: "1 flag de Roblox"})
	case flags > 1:
		out = append(out, ActiveChange{ID: "flags", Title: fmt.Sprintf("%d flags de Roblox", flags)})
	}
	return out
}
