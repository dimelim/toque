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

Elegir ciudad no usa ninguna API privada ni la cookie de la cuenta. Toque entra en servidores públicos y corta la entrada en cuanto el log muestra que el servidor es de otra ciudad, o si Roblox la deja en la fila de un servidor lleno. La lista pública de servidores solo admite 3 consultas por minuto, así que Toque hace una por búsqueda y espera si Roblox lo pide.

Roblox abre un servidor privado vacío cerca de quien entra primero. Para abrirlo en otra ciudad, el usuario enciende una VPN de ese país con su propia app. Toque no configura ninguna VPN, solo comprueba por qué país sale la conexión y en qué centro de datos se abrió el servidor.

## Cómo deshace los cambios

Cada optimización guarda en `C:\ProgramData\Toque\journal.json` el estado que había antes y solo después aplica el cambio, así que un cierre a mitad nunca deja algo aplicado sin forma de deshacerlo. El diario se escribe con un archivo temporal, se vuelca al disco y se renombra, y si aparece dañado Toque avisa en lugar de sobrescribirlo. Un experimento deja una marca mientras dura y, si se interrumpe, Toque devuelve el cambio a como estaba al volver a abrirse. Cuando Roblox se actualiza borra sus flags, y Toque las vuelve a poner al arrancar. Si el usuario cambia desde el menú de Roblox un ajuste que había puesto Toque, Toque lo da por suyo y no lo vuelve a imponer.

## Qué protege

Toque necesita permisos de administrador para PresentMon, pktmon y los ajustes de Windows. Las escrituras en las carpetas de Roblox, que son del usuario, se hacen con el token normal del usuario, de modo que un enlace o junction plantado ahí no puede desviarlas a una carpeta protegida. `C:\ProgramData\Toque` pertenece a Administradores y los usuarios solo pueden leerlo.

Las actualizaciones se publican en las versiones de GitHub de este repositorio con un manifiesto firmado con Ed25519. Toque comprueba la firma y el SHA-256 del ejecutable antes de instalarlo, y nunca lo hace a mitad de una prueba.
