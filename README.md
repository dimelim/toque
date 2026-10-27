# Toque

Aplicación de Windows que mide el retraso real de Roblox (enfocada en Realistic Street Soccer) y aplica solo los cambios que demuestran reducirlo con pruebas A/B. También te deja elegir la ciudad del servidor en el que juegas.

Es código abierto bajo [licencia MIT](LICENSE). Descarga la última versión firmada en [dimelim/toque-releases](https://github.com/dimelim/toque-releases/releases/latest).

## Cómo mide

- **Frames y latencia de render.** Intel PresentMon 2.6 lee los eventos ETW de DXGI/DxgKrnl del proceso de Roblox, sin inyectar nada. Se calcula FPS medio, 1 % y 0,1 % low, tirones (frames que superan 2,5 veces la mediana local y más de 8 ms), variación del frametime, modo de presentación y latencia CPU→pantalla.
- **Ping real del juego.** pktmon (incluido en Windows) captura el tráfico UDP con el servidor UDMUX de Roblox. El contenido va cifrado, pero los paquetes de confirmación tienen tamaños fijos (54 bytes de bajada, 56 de subida). El ping sale del pico del histograma "confirmación del servidor menos envío del cliente", restando el tick de confirmación del servidor (10,3 ms). Coincide con el NetworkPing que muestra Roblox en su panel de estadísticas.
- **Pérdida de paquetes de bajada.** Las confirmaciones del servidor llevan su contador de paquetes en el byte 36. Los huecos en la secuencia son paquetes perdidos.
- **Sistema.** Contadores PDH para CPU, GPU por proceso (`GPU Engine`, `engtype_3D`), lecturas de paginación y latencia de lectura por disco.
- **Carga.** El log de Roblox da la IP del servidor, el JobId y el tiempo de unión.

## Cómo decide

`internal/bench/ab.go` alterna sin el cambio (A) y con el cambio (B) en orden ABBA. Los cambios "en vivo" se alternan dentro de una misma sesión y los que requieren reabrir Roblox se prueban reentrando al mismo servidor por JobId. Cada métrica se compara por mediana de ventanas con intervalo de confianza bootstrap al 95 % y Mann-Whitney. Para las métricas principales el umbral se corrige por Bonferroni (0,05 dividido entre el número de métricas principales) y las secundarias solo cuentan como empeoramiento con p < 0,01. Las ventanas en las que Roblox pierde el foco o hay entrada del usuario se descartan.

Todo lo que pasó una primera prueba se repitió antes de darlo por bueno. Dos resultados iniciales "significativos" (el temporizador a 0,5 ms y el plan de energía) no se confirmaron en la réplica y se descartaron. Al juntar los datos de varias pruebas el umbral sube a p < 0,01.

Los resultados completos están en [docs/BENCHMARKS.md](docs/BENCHMARKS.md).

## Cómo deshace los cambios

Cada optimización guarda en `C:\ProgramData\Toque\journal.json` el estado que había antes y solo después aplica el cambio, así que un cierre a mitad nunca deja algo aplicado sin forma de deshacerlo. El diario se escribe con un archivo temporal, se vuelca al disco y se renombra, y si aparece dañado Toque avisa en lugar de sobrescribirlo. Un experimento deja una marca mientras dura y, si se interrumpe, Toque devuelve el cambio a como estaba al volver a abrirse. Cuando Roblox se actualiza borra sus flags, y Toque las vuelve a poner al arrancar. Si cambias desde el menú de Roblox un ajuste que había puesto Toque, Toque lo da por tuyo y no lo vuelve a imponer.

## Qué protege

Toque necesita permisos de administrador para PresentMon, pktmon y los ajustes de Windows. Las escrituras en las carpetas de Roblox, que son del usuario, se hacen con el token normal del usuario, de modo que un enlace o junction plantado ahí no puede desviarlas a una carpeta protegida. `C:\ProgramData\Toque` pertenece a Administradores y los usuarios solo pueden leerlo. Solo se usan flags de la lista oficial que Roblox permite, nunca se lee la cookie de la cuenta ni se toca la memoria del juego.

Roblox escribe en su log el número de centro de datos de cada servidor. Toque lo traduce a una ciudad con la lista pública de [RoValra](https://github.com/NotValra/RoValra) (`apis.rovalra.com`), que descarga como mucho una vez al día y guarda en `C:\ProgramData\Toque\dcmap.json`. Para los servidores que no salen en esa lista y para ordenar las ciudades por cercanía, consulta ipinfo.io y guarda las respuestas en `C:\ProgramData\Toque\places.json`. Elegir región no usa ninguna API privada. Toque entra en servidores públicos y corta la entrada en cuanto el log de Roblox muestra que el servidor es de otra ciudad, o si Roblox la deja en la fila de un servidor lleno. El enlace de un servidor privado se guarda solo en este PC y lo abre el propio Roblox, sin pasar por ningún servicio de Toque.

## Estructura

| Ruta | Qué hace |
|---|---|
| `main.go`, `cli.go` | App Wails (ventana, enlaces con el frontend) y modos `--uninstall` / `--revert-all` |
| `cmd/toque-lab` | CLI de laboratorio para experimentos y el informe de benchmarks |
| `internal/roblox` | Instalación, log, configuración, FFlags permitidas, unión por enlace profundo, servidores privados, lista de servidores |
| `internal/frames` | PresentMon y resumen de frames |
| `internal/netq` | pktmon, lector pcapng, estimador de ping, traceroute ICMP |
| `internal/sysmon` | Muestreo PDH de CPU, GPU, memoria y disco |
| `internal/tweaks` | Cada optimización con aplicar, revertir y estado, más el diario que permite deshacerlo todo |
| `internal/catalog` | Qué cambia cada optimización, por qué, su riesgo y el veredicto de las pruebas |
| `internal/bench` | Medición de ventanas, experimentos A/B, carga artificial |
| `internal/app` | Servicio que usa la interfaz (diagnóstico, catálogo, pruebas, jugar cerca o en una región, instalador) |
| `internal/geo` | Ciudades de los centros de datos de Roblox (RoValra) y de una IP (ipinfo.io) |
| `internal/install` | Autoinstalación en Program Files, accesos directos, desinstalación |
| `internal/update` | Comprueba, descarga y verifica las versiones nuevas firmadas |
| `cmd/toque-sign` | Genera la clave de firma y firma el manifiesto de cada versión |
| `frontend` | React 19, Tailwind 4, componentes de BoardUI, iconos Gravity UI |

## Compilar

Hace falta Go 1.27, Node 24 y Wails CLI 2.17. Toque incluye el `PresentMon.exe` de consola que publica Intel en [GameTechDev/PresentMon](https://github.com/GameTechDev/PresentMon/releases). No está en el repositorio, así que descárgalo y pásalo al guion, que comprueba que lleve la firma de Intel.

```powershell
.\build.ps1 -PresentMon C:\ruta\a\PresentMon.exe
```

El resultado es un único `build\bin\Toque.exe` que se instala solo al abrirlo.

## Publicar una versión

Sube `Version` en `internal/app/service.go`, haz commit y ejecuta `.\release.ps1 -Notes "..."`. El guion compila, firma `latest.json` con la clave de `%USERPROFILE%\.toque\update-signing.key` y publica el ejecutable en [dimelim/toque-releases](https://github.com/dimelim/toque-releases), que es público y solo tiene binarios, y en este repositorio.

Las copias instaladas consultan el manifiesto al abrirse y cada 6 horas. Si hay una versión nueva la descargan, comprueban la firma Ed25519 y el SHA-256 y la instalan al volver a abrir Toque o al pulsar Actualizar, nunca a mitad de una prueba. La clave privada no está en el repositorio. Si se pierde, las copias instaladas dejarán de aceptar versiones nuevas hasta que se instale a mano una que lleve la clave nueva. Se compila sin símbolos ni rutas del equipo de desarrollo. La ofuscación con garble se probó y se descartó porque Microsoft Defender marca el binario ofuscado como troyano (Trojan:Win32/Gracing.I), mientras que el binario normal sale limpio.

## Licencia

MIT, ver [LICENSE](LICENSE). El software de terceros que va dentro conserva su propia licencia y aparece en [assets/THIRD-PARTY-NOTICES.txt](assets/THIRD-PARTY-NOTICES.txt). Toque no está afiliado a Roblox Corporation.
