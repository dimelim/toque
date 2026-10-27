# Toque

Aplicación de Windows que mide el retraso real de Roblox (enfocada en Realistic Street Soccer) y aplica solo los cambios que demuestran reducirlo con pruebas A/B.

## Cómo mide

- **Frames y latencia de render.** Intel PresentMon 2.6 lee los eventos ETW de DXGI/DxgKrnl del proceso de Roblox, sin inyectar nada. Se calcula FPS medio, 1 % y 0,1 % low, tirones (frames que superan 2,5 veces la mediana local y más de 8 ms), variación del frametime, modo de presentación y latencia CPU→pantalla.
- **Ping real del juego.** pktmon (incluido en Windows) captura el tráfico UDP con el servidor UDMUX de Roblox. El contenido va cifrado, pero los paquetes de confirmación tienen tamaños fijos (54 bytes de bajada, 56 de subida). El ping sale del pico del histograma "confirmación del servidor menos envío del cliente", restando el tick de confirmación del servidor (10,3 ms). Coincide con el NetworkPing que muestra Roblox en su panel de estadísticas.
- **Pérdida de paquetes de bajada.** Las confirmaciones del servidor llevan su contador de paquetes en el byte 36. Los huecos en la secuencia son paquetes perdidos.
- **Sistema.** Contadores PDH para CPU, GPU por proceso (`GPU Engine`, `engtype_3D`), lecturas de paginación y latencia de lectura por disco.
- **Carga.** El log de Roblox da la IP del servidor, el JobId y el tiempo de unión.

## Cómo decide

`internal/bench/ab.go` alterna sin el cambio (A) y con el cambio (B) en orden ABBA. Los cambios "en vivo" se alternan dentro de una misma sesión y los que requieren reabrir Roblox se prueban reentrando al mismo servidor por JobId. Cada métrica se compara por mediana de ventanas con intervalo de confianza bootstrap al 95 % y Mann-Whitney. Solo es concluyente si el intervalo no cruza el cero y p < 0,05. Las ventanas en las que Roblox pierde el foco o hay entrada del usuario se descartan.

Los resultados completos están en [docs/BENCHMARKS.md](docs/BENCHMARKS.md).

## Estructura

| Ruta | Qué hace |
|---|---|
| `main.go`, `cli.go` | App Wails (ventana, enlaces con el frontend) y modos `--uninstall` / `--revert-all` |
| `cmd/toque-lab` | CLI de laboratorio para experimentos y el informe de benchmarks |
| `internal/roblox` | Instalación, log, configuración, FFlags permitidas, unión por enlace profundo, lista de servidores |
| `internal/frames` | PresentMon y resumen de frames |
| `internal/netq` | pktmon, lector pcapng, estimador de ping, traceroute ICMP |
| `internal/sysmon` | Muestreo PDH de CPU, GPU, memoria y disco |
| `internal/tweaks` | Cada optimización con aplicar, revertir y estado, más el diario que permite deshacerlo todo |
| `internal/catalog` | Qué cambia cada optimización, por qué, su riesgo y el veredicto de las pruebas |
| `internal/bench` | Medición de ventanas, experimentos A/B, carga artificial |
| `internal/app` | Servicio que usa la interfaz (diagnóstico, catálogo, pruebas, buscador de servidor, instalador) |
| `internal/install` | Autoinstalación en Program Files, accesos directos, desinstalación |
| `frontend` | React 19, Tailwind 4, componentes de BoardUI, iconos Gravity UI |

## Compilar

Hace falta Go 1.27, Node 24, Wails CLI 2.17 y, para la versión ofuscada, garble 0.18.

```powershell
.\build.ps1             # binario normal
.\build.ps1 -Obfuscate  # con garble (-literals -tiny)
```

El resultado es un único `build\bin\Toque.exe` que se instala solo al abrirlo.
