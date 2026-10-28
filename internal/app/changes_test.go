package app

import "testing"

func TestChangeTitlesAreHuman(t *testing.T) {
	cases := map[string]string{
		"privacy:ads":                         "Publicidad personalizada",
		"windowed-games":                      "Juegos en ventana",
		"windows:gamemode":                    "Modo de juego de Windows",
		"setting:FramerateCap=144":            "Límite de FPS a 144",
		"setting:SavedQualityLevel=9":         "Gráficos en nivel 9 de 10",
		"flag:DFFlagDebugPauseVoxelizer=True": "Iluminación estática",
	}
	for id, want := range cases {
		if got := changeTitle(id); got != want {
			t.Errorf("changeTitle(%q) = %q, want %q", id, got, want)
		}
	}
}
