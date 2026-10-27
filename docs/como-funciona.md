# Cómo funciona Toque

## Cómo mide

- **Frames y latencia de render.** Intel PresentMon lee los eventos ETW de DXGI y DxgKrnl del proceso de Roblox, sin inyectar nada. Con ellos calcula FPS medio, 1 % y 0,1 % low, tirones (frames que superan 2,5 veces la mediana local y más de 8 ms), variación del frametime, modo de presentación y latencia de la CPU a la pantalla.
- **Ping real del juego.** pktmon, incluido en Windows, captura el tráfico UDP con el servidor UDMUX de Roblox. El contenido va cifrado, pero los paquetes de confirmación tienen tamaños fijos, 54 bytes de bajada y 56 de subida. El ping sale del pico del histograma "confirmación del servidor menos envío del cliente", restando el tick con el que el cliente agrupa sus confirmaciones. Coincide con el NetworkPing que muestra Roblox en su panel de estadísticas.
- **Pérdida de paquetes de bajada.** Las confirmaciones del servidor llevan su contador de paquetes en el byte 36. Los huecos en la secuencia son paquetes perdidos.
- **Sistema.** Contadores PDH para CPU, GPU por proceso (`GPU Engine`, `engtype_3D`), lecturas de paginación y latencia de lectura por disco.
- **Servidor.** El log de Roblox da la IP del servidor, el JobId, el número de centro de datos y el tiempo de unión.

## Cómo decide

`internal/bench/ab.go` alterna sin el cambio (A) y con el cambio (B) en orden ABBA. Los cambios que se aplican en vivo se alternan dentro de una misma sesión, y los que requieren reabrir Roblox se prueban reentrando al mismo servidor por JobId. Cada métrica se compara por mediana de ventanas, con intervalo de confianza bootstrap al 95 % y prueba de Mann-Whitney. Para las métricas principales el umbral se corrige por Bonferroni, y las secundarias solo cuentan como empeoramiento con p < 0,01. Se descartan las ventanas en las que Roblox pierde el foco o hay entrada del usuario.

Todo lo que pasa una primera prueba se repite antes de darlo por bueno, y al juntar los datos de varias pruebas el umbral sube a p < 0,01. Cada optimización muestra su veredicto en la pantalla Optimizar, y cualquiera puede repetir la prueba en su PC con «Probar en mi PC».

## Cómo elige servidor

Roblox escribe en su log el número de centro de datos de cada servidor. Toque lo traduce a una ciudad con la lista pública de RoValra (`apis.rovalra.com`), que descarga como mucho una vez al día. Para los servidores que no están en esa lista y para ordenar las ciudades por cercanía consulta ipinfo.io.

Elegir ciudad no usa ninguna API privada ni la cookie de la cuenta. Toque lee la lista pública de servidores de Roblox y pregunta a la API pública de RoValra (`/v1/servers/details`) en qué centro de datos está cada uno. RoValra lo sabe porque sus usuarios lo reportan al ver la lista de servidores, así que conoce más o menos un tercio de ellos. Si alguno con sitio está en la ciudad elegida, Toque entra directo y comprueba en el log que de verdad es de esa ciudad. Si Roblox deja la entrada en la fila de un servidor lleno, pasa al siguiente.

Si no hay ninguno, el vigía revisa la lista de menos a más llena, una página cada 25 segundos para no pasar el límite de Roblox de 3 consultas por minuto, y entra en cuanto aparece una partida de esa ciudad, aunque estés jugando en otra. Se detiene a la hora.

Roblox abre un servidor privado vacío en el centro de datos que elige para quien entra primero, cerca de su conexión y según el sitio libre que tenga. Toque no puede cambiar esa decisión, y los centros nuevos como Lima, Santiago o Querétaro casi no reciben partidas. Si eliges una ciudad en «Abrir en», Toque entra y lee en el log en qué ciudad quedó y cuándo arrancó el servidor. Así distingue si ya estaba abierto o si Roblox eligió otra ciudad al abrirlo, y te dice qué hacer. Para que se abra en otra ciudad tiene que estar vacío y tiene que entrar primero alguien que se conecte desde allí, o con una VPN de allí. Toque no configura ninguna VPN.

## Cómo aplica las flags

Toque escribe las flags en `ClientSettings\ClientAppSettings.json`, dentro de la carpeta de la versión de Roblox, igual que Bloxstrap y Fishstrap. Acepta cualquier flag de cualquier tipo, y los ajustes más usados tienen controles propios. Cada cambio entra en el diario igual que una optimización, y Toque lo vuelve a poner cuando una actualización de Roblox borra la carpeta.

Desde septiembre de 2025 Roblox solo hace caso a las flags de su lista oficial. Toque nunca escribe flags en la memoria de Roblox para saltarse esa lista, porque el anticheat lo trata como una trampa que puede acabar en baneo.

## El HUD

El HUD es una ventana propia, siempre encima y transparente a los clics, que Toque dibuja con GDI sobre Roblox. Para saber qué teclas pulsas no engancha el teclado, solo mira el estado de W, A, S, D, espacio, shift y los botones del ratón con `GetAsyncKeyState` sesenta veces por segundo, así que no registra lo que escribes ni se parece a un keylogger. El ping es el del servidor en el que estás, o el de tu mejor centro cuando no hay partida. Se ve con Roblox en ventana o en pantalla completa sin bordes, no en pantalla completa exclusiva, y su posición, tamaño y opacidad se guardan en `hud.json`.

## Cómo deshace los cambios

Cada optimización guarda en `C:\ProgramData\Toque\journal.json` el estado que había antes y solo después aplica el cambio, así que un cierre a mitad nunca deja algo aplicado sin forma de deshacerlo. El diario se escribe con un archivo temporal, se vuelca al disco y se renombra, y si aparece dañado Toque avisa en lugar de sobrescribirlo. Un experimento deja una marca mientras dura y, si se interrumpe, Toque devuelve el cambio a como estaba al volver a abrirse. Cuando Roblox se actualiza borra sus flags, y Toque las vuelve a poner al arrancar. Si el usuario cambia desde el menú de Roblox un ajuste que había puesto Toque, Toque lo da por suyo y no lo vuelve a imponer.

## Qué protege

Toque necesita permisos de administrador para PresentMon, pktmon y los ajustes de Windows. Las escrituras en las carpetas de Roblox, que son del usuario, se hacen con el token normal del usuario, de modo que un enlace o junction plantado ahí no puede desviarlas a una carpeta protegida. `C:\ProgramData\Toque` pertenece a Administradores y los usuarios solo pueden leerlo.

Las actualizaciones se publican en las versiones de GitHub de este repositorio con un manifiesto firmado con Ed25519. Toque comprueba la firma y el SHA-256 del ejecutable antes de instalarlo, y nunca lo hace a mitad de una prueba.
