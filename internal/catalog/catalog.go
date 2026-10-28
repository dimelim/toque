package catalog

import (
	"strings"

	"toque/internal/i18n"
)

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

	CategorySession = "sesion"
)

var Catalog = []Entry{
	{
		ID:       "pagefile-ssd",
		Title:    i18n.Mark("Archivo de paginación en el SSD"),
		Category: "memoria",
		What:     i18n.Mark("Mueve el archivo de paginación de Windows del disco mecánico al SSD del sistema, con un tamaño de 4 a 12 GB."),
		Why:      i18n.Mark("Cuando la RAM se llena, Windows guarda páginas en ese archivo y las vuelve a leer en mitad de la partida. En un disco mecánico cada lectura tarda entre 9 y 70 ms, suficiente para congelar un frame. En un SSD tarda alrededor de 1 ms."),
		Metric:   i18n.Mark("Tirones por minuto, 0,1 % low y latencia de lectura de la paginación."),
		Risk:     i18n.Mark("Ocupa entre 4 y 12 GB del SSD. Si quedan menos de 14 GB libres, Toque no lo aplica."),
		Revert:   i18n.Mark("Restaura la configuración anterior del archivo de paginación. Necesita reiniciar."),
		Verdict:  VerdictSituational,
		Lab:      i18n.Mark("Solo si el archivo está en un disco mecánico. En las pruebas cada lectura de paginación pasó de 9-21 ms a 0,3-1,1 ms, y la mediana del 1 % low subió de 104 a 122 FPS con menos RAM libre que antes."),
	},
	{
		ID:       "ifeo-priority-high",
		Title:    i18n.Mark("Prioridad alta para Roblox"),
		Category: "sistema",
		What:     i18n.Mark("Le dice a Windows que arranque Roblox siempre con prioridad de CPU alta, desde el registro del propio sistema."),
		Why:      i18n.Mark("Si otro programa compite por la CPU, Windows atiende antes a Roblox y el juego pierde menos tiempo esperando turno."),
		Metric:   i18n.Mark("1 % low y tirones con carga de fondo."),
		Risk:     i18n.Mark("Los programas de fondo pueden ir más lentos mientras juegas."),
		Revert:   i18n.Mark("Borra la entrada de registro. Se aplica al siguiente arranque de Roblox."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada tras tres pruebas con 10 hilos de carga de fondo (25 ventanas por lado). Las dos primeras apuntaban a mejor 1 % low sin llegar a ser concluyentes, y la réplica final salió al revés. Juntando todo, +3,9 %, con p = 0,52."),
	},
	{
		ID:       "proc-affinity-skip-core0",
		Title:    i18n.Mark("Roblox fuera del núcleo 0 de la CPU"),
		Category: "sistema",
		What:     i18n.Mark("Quita a Roblox los dos primeros hilos de la CPU, donde Windows atiende buena parte de las interrupciones de la red, el disco y la gráfica."),
		Why:      i18n.Mark("La idea es que el hilo principal del juego no compita con esas interrupciones y entregue frames más regulares."),
		Metric:   i18n.Mark("1 % low, tirones y latencia de render."),
		Risk:     i18n.Mark("Roblox tiene dos hilos menos para repartir su trabajo."),
		Revert:   i18n.Mark("Devuelve a Roblox todos los hilos. También se pierde al cerrar el juego."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada. En 12 ventanas, con el cambio aplicado en vivo, la latencia de render quedó igual (6,98 frente a 6,99 ms) y los tirones también (10,8 por minuto en los dos casos). El 1 % low varió mucho entre ventanas y no mostró un efecto claro."),
	},
	{
		ID:       "power-plan:8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c",
		Title:    i18n.Mark("Plan de energía de alto rendimiento"),
		Category: "sistema",
		What:     i18n.Mark("Cambia el plan de energía de Windows a Alto rendimiento."),
		Why:      i18n.Mark("El plan Equilibrado baja la frecuencia del procesador y aparca núcleos cuando cree que no hacen falta, y volver a subir cuesta milisegundos en mitad de un frame."),
		Metric:   i18n.Mark("1 % low, tirones y latencia de render."),
		Risk:     i18n.Mark("Más consumo y temperatura en reposo."),
		Revert:   i18n.Mark("Vuelve al plan que tenías antes."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada en las pruebas, hechas en un PC de escritorio. Alto rendimiento frente a Equilibrado no cambió el 1 % low, los tirones ni la latencia en 12 ventanas. En portátiles puede ser distinto, por eso puedes probarla en el tuyo."),
	},
	{
		ID:       "timer-res-0.5ms",
		Title:    i18n.Mark("Temporizador de Windows a 0,5 ms"),
		Category: "sistema",
		What:     i18n.Mark("Pide al sistema una resolución de temporizador de 0,5 ms mientras juegas."),
		Why:      i18n.Mark("El limitador de FPS de Roblox duerme entre frames y la precisión de ese sueño depende del temporizador del sistema."),
		Metric:   i18n.Mark("Estabilidad del frametime y 1 % low."),
		Risk:     i18n.Mark("Algo más de consumo de energía."),
		Revert:   i18n.Mark("Se libera al cerrar Roblox o al desactivarla."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada tras tres pruebas con 25 ventanas por lado. La primera subió el 1 % low un 18 %, pero las dos réplicas no lo confirmaron, y la última incluso salió al revés. Juntando todo, +1,3 %, con p = 0,47. Fue un falso positivo."),
	},
	{
		ID:       "flag:FFlagDebugGraphicsPreferD3D11=True",
		Title:    i18n.Mark("Forzar Direct3D 11"),
		Category: "fps",
		What:     i18n.Mark("Obliga a Roblox a dibujar con Direct3D 11, con una opción oficial de su lista permitida."),
		Why:      i18n.Mark("Con Direct3D 11 Roblox entrega cada frame directo a la pantalla. Con Vulkan pasa por el compositor de Windows y cada frame tarda más en verse."),
		Metric:   i18n.Mark("Latencia de render."),
		Risk:     i18n.Mark("Ninguno conocido. Es el modo por defecto en la mayoría de equipos."),
		Revert:   i18n.Mark("Quita la opción del archivo de configuración de Roblox."),
		Verdict:  VerdictSituational,
		Lab:      i18n.Mark("Solo ayuda si el diagnóstico ve que Roblox pasa por el compositor. En las pruebas Direct3D 11 tuvo 7,4 ms de latencia de render frente a 18,8 ms con Vulkan."),
	},
	{
		ID:       "flag:FFlagDebugGraphicsPreferVulkan=True",
		Title:    i18n.Mark("Render con Vulkan"),
		Category: "fps",
		What:     i18n.Mark("Hace que Roblox dibuje con Vulkan en lugar de Direct3D 11, usando una opción oficial de su lista permitida."),
		Why:      i18n.Mark("Según la GPU y el driver, una API gasta menos CPU por frame que la otra."),
		Metric:   i18n.Mark("FPS, 1 % low y tiempo de CPU por frame."),
		Risk:     i18n.Mark("Con Vulkan, Roblox pasa por el compositor de Windows y cada frame tarda más en llegar a la pantalla."),
		Revert:   i18n.Mark("Quita la opción del archivo de configuración de Roblox."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada. Más que duplicó la latencia de render (7,4 a 18,8 ms, +154 %) sin mejorar los FPS ni los tirones de forma medible."),
	},
	{
		ID:       "flag:DFFlagDebugPauseVoxelizer=True",
		Title:    i18n.Mark("Iluminación estática"),
		Category: "fps",
		What:     i18n.Mark("Pausa el voxelizador que recalcula la iluminación del mapa, con una opción oficial de la lista permitida."),
		Why:      i18n.Mark("Ese cálculo consume CPU y GPU en cada frame aunque el mapa no cambie."),
		Metric:   i18n.Mark("Latencia de render y tiempo de GPU por frame."),
		Risk:     i18n.Mark("La iluminación del mapa deja de recalcularse, y algunos efectos de luz dinámicos pueden verse menos realistas."),
		Revert:   i18n.Mark("Quita la opción del archivo de configuración de Roblox."),
		Verdict:  VerdictProven,
		Lab:      i18n.Mark("Probada y replicada. En la réplica, con 6 reaperturas por lado, la latencia de render bajó un 4,4 % (7,08 a 6,77 ms, p = 0,008) y la GPU por frame un 6 % (p = 0,020). La primera prueba ya había dado −3,6 y −5,3 %. Es una mejora pequeña pero constante."),
	},
	{
		ID:       "setting:SavedQualityLevel=9",
		Title:    i18n.Mark("Gráficos en nivel 9 de 10"),
		Category: "fps",
		What:     i18n.Mark("Fija la calidad gráfica de Roblox en el nivel 9 de 10 en lugar de dejarla en automático."),
		Why:      i18n.Mark("En este PC el modo automático rinde igual que el nivel 10, el máximo, y es ese último nivel el que dispara el trabajo de la GPU. Los niveles 1, 4, 6, 8 y 9 dejaron la latencia y la GPU por frame igual de bajas, así que el 9 conserva todo el ahorro y es el que mejor se ve."),
		Metric:   i18n.Mark("Latencia de render y tiempo de GPU por frame."),
		Risk:     i18n.Mark("Se ve un poco peor que en el nivel máximo. En un PC más modesto el automático puede estar ya por debajo del 9 y entonces no ayudaría, por eso conviene comprobarlo con \"Probar en mi PC\"."),
		Revert:   i18n.Mark("Vuelve a la calidad que tenías antes en la configuración de Roblox."),
		Verdict:  VerdictTradeoff,
		Lab:      i18n.Mark("Probada y replicada frente a la calidad automática, con 12 reaperturas por prueba. Con el nivel 9 la latencia de render bajó un 12 a 13 % (de 7,1 a 6,2 ms) y la GPU por frame un 29 a 31 %, en todas las reaperturas. Los niveles 1, 4, 6 y 8 dieron la misma mejora y además unos 60 a 90 MB menos de RAM, y el 10 no cambió nada porque en las pruebas equivalía al automático."),
	},
	{
		ID:       "flag:FIntDebugForceMSAASamples=0",
		Title:    i18n.Mark("Sin antialiasing (MSAA)"),
		Category: "fps",
		What:     i18n.Mark("Desactiva el suavizado de bordes MSAA de Roblox, con una opción oficial de su lista permitida."),
		Why:      i18n.Mark("El MSAA dibuja varias muestras por píxel en los bordes, y eso cuesta tiempo de GPU en cada frame."),
		Metric:   i18n.Mark("Tiempo de GPU por frame y latencia de render."),
		Risk:     i18n.Mark("Los bordes de los objetos se ven con dientes de sierra."),
		Revert:   i18n.Mark("Quita la opción del archivo de configuración de Roblox."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada. La GPU por frame quedó igual (2,60 frente a 2,62 ms) en 12 reaperturas, porque con la calidad automática Roblox no estaba usando MSAA en las pruebas."),
	},
	{
		ID:       "flag:FIntFRMMinGrassDistance=0,FIntFRMMaxGrassDistance=0",
		Title:    i18n.Mark("Sin hierba 3D"),
		Category: "fps",
		What:     i18n.Mark("Desactiva la hierba en relieve del terreno, con opciones oficiales de la lista permitida."),
		Why:      i18n.Mark("La hierba del terreno se dibuja con miles de briznas, y eso cuesta GPU y CPU en mapas abiertos."),
		Metric:   i18n.Mark("FPS, 1 % low y tiempo de GPU por frame."),
		Risk:     i18n.Mark("El césped se ve plano en mapas con terreno."),
		Revert:   i18n.Mark("Quita las opciones del archivo de configuración de Roblox."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada para Realistic Street Soccer. Su campo no usa hierba de terreno y la GPU por frame quedó igual (2,57 frente a 2,58 ms)."),
	},
	{
		ID:       "flag:DFIntCSGLevelOfDetailSwitchingDistance=0,DFIntCSGLevelOfDetailSwitchingDistanceL12=0,DFIntCSGLevelOfDetailSwitchingDistanceL23=0,DFIntCSGLevelOfDetailSwitchingDistanceL34=0",
		Title:    i18n.Mark("Piezas compuestas simplificadas"),
		Category: "fps",
		What:     i18n.Mark("Hace que Roblox dibuje siempre la versión más simple de las piezas compuestas (uniones y cortes), a cualquier distancia, con opciones oficiales de la lista permitida."),
		Why:      i18n.Mark("Esas piezas tienen varias versiones con menos polígonos para verlas de lejos. Usar siempre la más simple ahorra trabajo de GPU en mapas que tienen muchas."),
		Metric:   i18n.Mark("Tiempo de GPU por frame y latencia de render."),
		Risk:     i18n.Mark("Las piezas compuestas se ven con menos detalle de cerca."),
		Revert:   i18n.Mark("Quita las opciones del archivo de configuración de Roblox."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada para Realistic Street Soccer. En 12 reaperturas la GPU por frame bajó un 1,8 % (2,61 a 2,56 ms) sin llegar al umbral de la prueba, y la latencia de render quedó igual (7,02 frente a 6,96 ms)."),
	},
	{
		ID:       "nic:*InterruptModeration=0",
		Title:    i18n.Mark("Tarjeta de red sin moderación de interrupciones"),
		Category: "red",
		What:     i18n.Mark("Desactiva la moderación de interrupciones de la tarjeta de red, que agrupa varios paquetes antes de avisar al procesador."),
		Why:      i18n.Mark("Es de los ajustes de red más recomendados para jugar, porque en teoría cada paquete del servidor llega al juego unas décimas de milisegundo antes."),
		Metric:   i18n.Mark("Ping real dentro del juego y jitter."),
		Risk:     i18n.Mark("Más interrupciones y algo más de uso de CPU cuando hay mucho tráfico. Al aplicarla la conexión se corta unos segundos mientras la tarjeta se reinicia."),
		Revert:   i18n.Mark("Vuelve a activar la moderación. La tarjeta se reinicia otra vez."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada. En 10 reaperturas válidas el ping real quedó igual (49,19 frente a 49,17 ms) y el jitter no cambió de forma medible. La diferencia teórica es menor que lo que varía el propio servidor."),
	},
	{
		ID:       "flag:DFFlagTextureQualityOverrideEnabled=True,DFIntTextureQualityOverride=1",
		Title:    i18n.Mark("Texturas ligeras"),
		Category: "memoria",
		What:     i18n.Mark("Fija la calidad de texturas de Roblox en un nivel bajo, con opciones oficiales de la lista permitida."),
		Why:      i18n.Mark("Las texturas grandes ocupan RAM y VRAM. En equipos de 8 GB eso empuja a Windows a paginar."),
		Metric:   i18n.Mark("RAM de Roblox, fallos de página duros y tirones."),
		Risk:     i18n.Mark("Las texturas se ven más borrosas de cerca."),
		Revert:   i18n.Mark("Quita las opciones del archivo de configuración de Roblox."),
		Verdict:  VerdictRejected,
		Lab:      i18n.Mark("Descartada. La RAM de Roblox quedó igual (1.954 frente a 1.946 MB) y no cambiaron los tirones ni las lecturas de paginación de forma medible."),
	},
	{
		ID:       "session:background-eco",
		Title:    i18n.Mark("Apps de fondo en modo eficiencia"),
		Category: CategorySession,
		What:     i18n.Mark("Mientras juegas, pone en modo eficiencia los programas conocidos que gastan CPU de fondo, como navegadores, launchers o fondos animados."),
		Why:      i18n.Mark("Un programa que gasta CPU de fondo compite con Roblox por el procesador. En modo eficiencia Windows lo atiende después que al juego y a menor frecuencia."),
		Metric:   i18n.Mark("1 % low, tirones y latencia de render con tus programas de fondo abiertos."),
		Risk:     i18n.Mark("Ese programa va más lento mientras juegas. Toque nunca toca los que están sonando ni la app que tengas delante, y lo devuelve a la normalidad al terminar."),
		Revert:   i18n.Mark("Cada programa vuelve a su prioridad al cerrar Roblox, al apagar la opción, al cerrar Toque o al volver a abrirlo tras un fallo."),
		Verdict:  VerdictPending,
		Lab:      i18n.Mark("Sin probar todavía. En el PC de pruebas, subir la prioridad de Roblox con carga de fondo no cambió nada medible, así que en equipos con muchos núcleos se espera poco efecto. Puede ayudar en equipos de 4 hilos con el navegador abierto."),
	},
	{
		ID:       "session:windows-fx",
		Title:    i18n.Mark("Windows en modo juego"),
		Category: CategorySession,
		What:     i18n.Mark("Mientras juegas, Windows deja de grabar en segundo plano y apaga animaciones, sombras y transparencia."),
		Why:      i18n.Mark("La grabación en segundo plano de la Xbox Game Bar codifica vídeo todo el tiempo, y Microsoft reconoce que cuesta rendimiento salvo en gráficas de gama alta. Las animaciones y la transparencia hacen trabajar al compositor de Windows cuando se ven."),
		Metric:   i18n.Mark("1 % low, tirones y latencia de render."),
		Risk:     i18n.Mark("Windows se ve más plano mientras juegas."),
		Revert:   i18n.Mark("Toque guarda tus valores antes de cambiarlos y los devuelve al cerrar Roblox, al apagar la opción, al cerrar Toque o al volver a abrirlo tras un fallo."),
		Verdict:  VerdictPending,
		Lab:      i18n.Mark("Sin probar todavía. Si la grabación en segundo plano ya estaba apagada, que es lo normal, el efecto esperado dentro de la partida es pequeño."),
	},
	{
		ID:       "session:power-overlay",
		Title:    i18n.Mark("Portátil en máximo rendimiento"),
		Category: CategorySession,
		What:     i18n.Mark("Mientras juegas con el portátil enchufado, pone el modo de energía de Windows en Máximo rendimiento."),
		Why:      i18n.Mark("En los portátiles, el modo de energía decide cuánto puede subir de frecuencia el procesador. El fabricante ajusta ese límite para cada modo."),
		Metric:   i18n.Mark("FPS, 1 % low y tirones."),
		Risk:     i18n.Mark("El portátil gasta más energía y se calienta más mientras juegas."),
		Revert:   i18n.Mark("Toque guarda tu modo antes de cambiarlo y lo devuelve al cerrar Roblox, al apagar la opción, al cerrar Toque o al volver a abrirlo tras un fallo."),
		Verdict:  VerdictPending,
		Lab:      i18n.Mark("Sin probar todavía. Es distinto del plan Alto rendimiento, que en el PC de sobremesa de pruebas no cambió nada. Solo se puede medir en un portátil."),
	},
}

var labTitles = map[string]string{
	"power-plan:381b4222-f694-41f0-9685-ff5bb260df2e": i18n.Mark("Plan Equilibrado frente al plan personalizado"),
	"proc-priority-high":                              i18n.Mark("Prioridad alta para Roblox (en vivo)"),
}

func (e Entry) Localized() Entry {
	e.Title = i18n.T(e.Title)
	e.What = i18n.T(e.What)
	e.Why = i18n.T(e.Why)
	e.Metric = i18n.T(e.Metric)
	e.Risk = i18n.T(e.Risk)
	e.Revert = i18n.T(e.Revert)
	e.Lab = i18n.T(e.Lab)
	return e
}

func Find(id string) (Entry, bool) {
	for _, e := range Catalog {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

func Title(id string) string {
	if e, ok := Find(id); ok {
		return i18n.T(e.Title)
	}
	if t, ok := labTitles[id]; ok {
		return i18n.T(t)
	}
	if level, ok := strings.CutPrefix(id, "setting:SavedQualityLevel="); ok {
		return i18n.Tf("Gráficos en nivel %s de 10", level)
	}
	return id
}
