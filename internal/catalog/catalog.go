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
	VerdictTradeoff    = "tradeoff"
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
		Verdict:  VerdictSituational,
		Lab:      "Solo si el archivo está en un disco mecánico. En las pruebas cada lectura de paginación pasó de 9-21 ms a 0,3-1,1 ms, y la mediana del 1 % low subió de 104 a 122 FPS con menos RAM libre que antes.",
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
		Verdict:  VerdictRejected,
		Lab:      "Descartada tras tres pruebas con 10 hilos de carga de fondo (25 ventanas por lado). Las dos primeras apuntaban a mejor 1 % low sin llegar a ser concluyentes, y la réplica final salió al revés. Juntando todo, +3,9 % con p = 0,52.",
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
		Verdict:  VerdictRejected,
		Lab:      "Descartada en las pruebas, hechas en un PC de escritorio. Alto rendimiento frente a Equilibrado no cambió el 1 % low, los tirones ni la latencia en 12 ventanas. En portátiles puede ser distinto, por eso puedes probarla en el tuyo.",
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
		Verdict:  VerdictRejected,
		Lab:      "Descartada tras tres pruebas con 25 ventanas por lado. La primera dio +18 % en el 1 % low, pero las dos réplicas no lo confirmaron, y la última incluso salió al revés. Juntando todo, +1,3 % con p = 0,47. Fue un falso positivo.",
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
		Lab:      "Probada y replicada. En la réplica, con 6 reaperturas por lado, la latencia de render bajó un 4,4 % (7,08 a 6,77 ms, p = 0,008) y la GPU por frame un 6 % (p = 0,020). La primera prueba ya había dado −3,6 % y −5,3 %. Es una mejora pequeña pero constante.",
	},
	{
		ID:       "setting:SavedQualityLevel=1",
		Title:    "Gráficos al mínimo",
		Category: "fps",
		What:     "Fija la calidad gráfica de Roblox en el nivel más bajo en lugar de dejarla en automático.",
		Why:      "En automático Roblox sube y baja el detalle sobre la marcha. Fijarlo al mínimo reduce el trabajo de la GPU y la CPU en cada frame.",
		Metric:   "Latencia de render, 1 % low y tiempo de GPU por frame.",
		Risk:     "El juego se ve bastante peor, con menos detalle, sombras y distancia de dibujado.",
		Revert:   "Vuelve a la calidad que tenías antes en la configuración de Roblox.",
		Verdict:  VerdictTradeoff,
		Lab:      "Probada y replicada frente a la calidad automática. En dos pruebas de 12 reaperturas cada una, la latencia de render bajó un 12 % (de unos 6,97 a 6,12 ms), la GPU por frame un 28 % y la RAM de Roblox unos 80 MB, en todas las reaperturas. A cambio el juego se ve peor.",
	},
	{
		ID:       "flag:FIntDebugForceMSAASamples=0",
		Title:    "Sin antialiasing (MSAA)",
		Category: "fps",
		What:     "Desactiva el suavizado de bordes MSAA de Roblox, con una opción oficial de su lista permitida.",
		Why:      "El MSAA dibuja varias muestras por píxel en los bordes, y eso cuesta tiempo de GPU en cada frame.",
		Metric:   "Tiempo de GPU por frame y latencia de render.",
		Risk:     "Los bordes de los objetos se ven con dientes de sierra.",
		Revert:   "Quita la opción del archivo de configuración de Roblox.",
		Verdict:  VerdictRejected,
		Lab:      "Descartada. La GPU por frame quedó igual (2,60 frente a 2,62 ms) en 12 reaperturas, porque con la calidad automática Roblox no estaba usando MSAA en las pruebas.",
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
