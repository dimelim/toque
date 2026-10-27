package catalog

type Entry struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	What     string `json:"what"`
	Why      string `json:"why"`
	Metric   string `json:"metric"`
	Risk     string `json:"risk"`
	Revert   string `json:"revert"`
	Verdict  string `json:"verdict"`
	Lab      string `json:"lab"`
}

const (
	VerdictProven      = "proven"
	VerdictRejected    = "rejected"
	VerdictPending     = "pending"
	VerdictSituational = "situational"
)

var Catalog = []Entry{
	{
		ID:       "pagefile-ssd",
		Title:    "Archivo de paginación en el SSD",
		Category: "memoria",
		What:     "Mueve el archivo de paginación de Windows del disco mecánico al SSD del sistema, con un tamaño de 4 a 12 GB.",
		Why:      "Cuando la RAM se llena, Windows guarda páginas en ese archivo y las vuelve a leer en mitad de la partida. En un disco mecánico cada lectura tarda entre 9 y 70 ms, suficiente para congelar un frame. En un SSD tarda alrededor de 1 ms.",
		Metric:   "Tirones por minuto, 0,1 % low y latencia de lectura de la paginación.",
		Risk:     "Ocupa entre 4 y 12 GB del SSD. Si quedan menos de 14 GB libres, Toque no lo aplica.",
		Revert:   "Restaura la configuración anterior del archivo de paginación. Necesita reiniciar.",
		Verdict:  VerdictPending,
	},
	{
		ID:       "ifeo-priority-high",
		Title:    "Prioridad alta para Roblox",
		Category: "sistema",
		What:     "Le dice a Windows que arranque Roblox siempre con prioridad de CPU alta, desde el registro del propio sistema.",
		Why:      "Si otro programa compite por la CPU, Windows atiende antes a Roblox y el juego pierde menos tiempo esperando turno.",
		Metric:   "1 % low y tirones con carga de fondo.",
		Risk:     "Los programas de fondo pueden ir más lentos mientras juegas.",
		Revert:   "Borra la entrada de registro. Se aplica al siguiente arranque de Roblox.",
		Verdict:  VerdictPending,
	},
	{
		ID:       "power-plan:8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c",
		Title:    "Plan de energía de alto rendimiento",
		Category: "sistema",
		What:     "Cambia el plan de energía de Windows a Alto rendimiento.",
		Why:      "El plan Equilibrado baja la frecuencia del procesador y aparca núcleos cuando cree que no hacen falta, y volver a subir cuesta milisegundos en mitad de un frame.",
		Metric:   "1 % low, tirones y latencia de render.",
		Risk:     "Más consumo y temperatura en reposo.",
		Revert:   "Vuelve al plan que tenías antes.",
		Verdict:  VerdictPending,
	},
	{
		ID:       "timer-res-0.5ms",
		Title:    "Temporizador de Windows a 0,5 ms",
		Category: "sistema",
		What:     "Pide al sistema una resolución de temporizador de 0,5 ms mientras juegas.",
		Why:      "El limitador de FPS de Roblox duerme entre frames y la precisión de ese sueño depende del temporizador del sistema.",
		Metric:   "Estabilidad del frametime y 1 % low.",
		Risk:     "Algo más de consumo de energía.",
		Revert:   "Se libera al cerrar Roblox o al desactivarla.",
		Verdict:  VerdictProven,
		Lab:      "En las pruebas, 1 % low de 96 a 114 FPS (+18 %) y variación del frametime −35 %, alternando 5 veces con y sin el cambio.",
	},
	{
		ID:       "flag:FFlagDebugGraphicsPreferD3D11=True",
		Title:    "Forzar Direct3D 11",
		Category: "fps",
		What:     "Obliga a Roblox a dibujar con Direct3D 11, con una opción oficial de su lista permitida.",
		Why:      "Con Direct3D 11 Roblox entrega cada frame directo a la pantalla. Con Vulkan pasa por el compositor de Windows y cada frame tarda más en verse.",
		Metric:   "Latencia de render.",
		Risk:     "Ninguno conocido. Es el modo por defecto en la mayoría de equipos.",
		Revert:   "Quita la opción del archivo de configuración de Roblox.",
		Verdict:  VerdictSituational,
		Lab:      "Solo ayuda si el diagnóstico ve que Roblox pasa por el compositor. En las pruebas Direct3D 11 tuvo 7,4 ms de latencia de render frente a 18,8 ms con Vulkan.",
	},
	{
		ID:       "flag:FFlagDebugGraphicsPreferVulkan=True",
		Title:    "Render con Vulkan",
		Category: "fps",
		What:     "Hace que Roblox dibuje con Vulkan en lugar de Direct3D 11, usando una opción oficial de su lista permitida.",
		Why:      "Según la GPU y el driver, una API gasta menos CPU por frame que la otra.",
		Metric:   "FPS, 1 % low y tiempo de CPU por frame.",
		Risk:     "Con Vulkan, Roblox pasa por el compositor de Windows y cada frame tarda más en llegar a la pantalla.",
		Revert:   "Quita la opción del archivo de configuración de Roblox.",
		Verdict:  VerdictRejected,
		Lab:      "Descartada. Más que duplicó la latencia de render (7,4 a 18,8 ms, +154 %) sin mejorar los FPS ni los tirones de forma medible.",
	},
	{
		ID:       "flag:DFFlagDebugPauseVoxelizer=True",
		Title:    "Iluminación estática",
		Category: "fps",
		What:     "Pausa el voxelizador que recalcula la iluminación del mapa, con una opción oficial de la lista permitida.",
		Why:      "Ese cálculo consume CPU y GPU en cada frame aunque el mapa no cambie.",
		Metric:   "Latencia de render y tiempo de GPU por frame.",
		Risk:     "La iluminación del mapa deja de recalcularse, y algunos efectos de luz dinámicos pueden verse menos realistas.",
		Revert:   "Quita la opción del archivo de configuración de Roblox.",
		Verdict:  VerdictProven,
		Lab:      "Mejora pequeña pero medible. Latencia de render −3,6 % (6,98 a 6,72 ms) y GPU por frame −5,3 %. Sin cambio medible en tirones.",
	},
	{
		ID:       "flag:FIntFRMMinGrassDistance=0,FIntFRMMaxGrassDistance=0",
		Title:    "Sin hierba 3D",
		Category: "fps",
		What:     "Desactiva la hierba en relieve del terreno, con opciones oficiales de la lista permitida.",
		Why:      "La hierba del terreno se dibuja con miles de briznas, y eso cuesta GPU y CPU en mapas abiertos.",
		Metric:   "FPS, 1 % low y tiempo de GPU por frame.",
		Risk:     "El césped se ve plano en mapas con terreno.",
		Revert:   "Quita las opciones del archivo de configuración de Roblox.",
		Verdict:  VerdictRejected,
		Lab:      "Descartada para Realistic Street Soccer. Su campo no usa hierba de terreno y la GPU por frame quedó igual (2,57 frente a 2,58 ms).",
	},
	{
		ID:       "flag:DFFlagTextureQualityOverrideEnabled=True,DFIntTextureQualityOverride=1",
		Title:    "Texturas ligeras",
		Category: "memoria",
		What:     "Fija la calidad de texturas de Roblox en un nivel bajo, con opciones oficiales de la lista permitida.",
		Why:      "Las texturas grandes ocupan RAM y VRAM. En equipos de 8 GB eso empuja a Windows a paginar.",
		Metric:   "RAM de Roblox, fallos de página duros y tirones.",
		Risk:     "Las texturas se ven más borrosas de cerca.",
		Revert:   "Quita las opciones del archivo de configuración de Roblox.",
		Verdict:  VerdictRejected,
		Lab:      "Descartada. La RAM de Roblox quedó igual (1.954 frente a 1.946 MB) y no cambiaron los tirones ni las lecturas de paginación de forma medible.",
	},
}

func Find(id string) (Entry, bool) {
	for _, e := range Catalog {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}
