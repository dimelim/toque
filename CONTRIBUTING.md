# Contribuir a Toque

Gracias por ayudar. Esta guía explica cómo compilar Toque, cómo probarlo y cómo proponer cambios.

## Requisitos

- Windows 11
- Go 1.27
- Node 24
- Wails CLI 2.17
- El `PresentMon.exe` de consola que publica Intel en [GameTechDev/PresentMon](https://github.com/GameTechDev/PresentMon/releases)

## Compilar

```powershell
.\build.ps1 -PresentMon C:\ruta\a\PresentMon.exe
```

El guion comprueba que PresentMon lleve la firma de Intel y deja un único `build\bin\Toque.exe` que se instala solo al abrirlo.

## Probar

```powershell
go vet ./...
go test ./internal/...
cd frontend
npx tsc --noEmit
```

Las optimizaciones se validan con pruebas A/B dentro del juego. `cmd/toque-lab` es la herramienta de laboratorio para lanzarlas, buscar servidores por ciudad y generar el informe de resultados en `dist/`.

## Estructura

| Ruta | Qué hace |
|---|---|
| `main.go`, `cli.go` | App Wails (ventana, enlaces con el frontend) y modos `--uninstall` y `--revert-all` |
| `cmd/toque-lab` | Herramienta de laboratorio para experimentos, búsquedas de servidor e informes |
| `cmd/toque-sign` | Genera la clave de firma y firma el manifiesto de cada versión |
| `internal/app` | Servicio que usa la interfaz (diagnóstico, catálogo, pruebas, jugar, instalador) |
| `internal/roblox` | Instalación, log, configuración, flags permitidas, unión por enlace profundo, servidores privados y públicos |
| `internal/bench` | Medición de ventanas, experimentos A/B y carga artificial |
| `internal/catalog` | Qué cambia cada optimización, por qué, su riesgo y el veredicto de las pruebas |
| `internal/tweaks` | Aplicar, revertir y comprobar cada optimización, con el diario que permite deshacerlo todo |
| `internal/frames` | PresentMon y resumen de frames |
| `internal/netq` | pktmon, lector pcapng, estimador de ping y traceroute |
| `internal/sysmon` | Muestreo de CPU, GPU, memoria y disco |
| `internal/geo` | Ciudades de los centros de datos de Roblox y de una IP |
| `internal/install` | Autoinstalación, accesos directos y desinstalación |
| `internal/update` | Comprueba, descarga y verifica las versiones firmadas |
| `frontend` | React 19, Tailwind 4, componentes de BoardUI, iconos de Gravity UI y gráficas de Recharts |

## Proponer cambios

1. Abre un issue para hablar del cambio si es grande.
2. Haz un fork y una rama para tu cambio.
3. Comprueba que `go vet`, las pruebas y la comprobación de tipos pasan.
4. Abre el pull request explicando qué cambia y cómo lo probaste.

Una optimización nueva solo entra si una prueba A/B dentro del juego demuestra que mejora sin empeorar otra métrica.

## Versiones

Solo el mantenedor publica versiones. El manifiesto de cada versión va firmado con una clave privada que no está en el repositorio, y Toque solo instala actualizaciones con esa firma.
