## Objetivo General

Diseñar y programar en C++ un **Remote Access Trojan (RAT)** minimalista que permita recibir y ejecutar órdenes remotas en Windows, devolviendo los resultados a través de un canal cifrado con **AES-128**. Además, debe implementar persistencia básica mediante una entrada en el Registro de Windows.

***

## Descripción del Reto

1.  **Canal de Comando Cifrado**
    * El cliente RAT debe conectarse a un servidor remoto (IP:port) y negociar un canal cifrado con una clave AES-128 precompartida.
    * Todo el tráfico de órdenes y respuestas pasará por este canal seguro.

2.  **Interpretación y Ejecución de Comandos**
    * El RAT recibe cadenas de texto (por ejemplo: `dir`, `whoami`, `hostname`) y ejecuta el comando en la shell de Windows.
    * Captura la salida estándar y de error, la cifra y la envía de vuelta al servidor.

3.  **Persistencia Básica**
    * Al iniciar, si se marca la opción, el cliente debe copiarse a `%APPDATA%` y crear/modificar una clave en el Registro (`HKCU\Software\<TuNombre>\AutoStart`) apuntando al ejecutable.
    * Debe soportar también la eliminación de la persistencia mediante un comando remoto.

4.  **Entorno Controlado**
    * Probar únicamente en máquinas virtuales aisladas con snapshots limpios para garantizar la seguridad.

***

## Investigación Sugerida

* **Sockets en Windows**: `Winsock2` (`<winsock2.h>`, `WSAStartup`, `socket`, `connect`, `send`, `recv`).
* **AES-128 en C++**: Usar `OpenSSL` (`<openssl/aes.h>`) o una biblioteca ligera como `Crypto++`.
* **Ejecución de procesos**: `CreateProcessA`, redirección de `stdout`/`stderr` con pipes.
* **Persistencia en Registro**: API de Windows (`RegCreateKeyExA`, `RegSetValueExA`, `RegDeleteKeyA`).
* **Serialización de mensajes**: Definir un formato simple (por ejemplo, `longitud + payload cifrado`).

***

## Entrada Esperada del Programa (Cliente RAT)

Al ejecutar el binario, el cliente debe aceptar parámetros por línea de comandos o desde un archivo de configuración:

1.  **IP del servidor** de control (string).
2.  **Puerto** de conexión (int).
3.  **Clave AES-128** en formato hexadecimal (16 bytes).
4.  **Flag de persistencia** (`--persist` o `--nopersist`).

***

## Salida Esperada del Programa

* **Consola (modo debug)**: Información sobre la conexión, el cifrado, y los comandos recibidos y enviados.
* **Servidor de control**: Al enviarle un comando, recibe la respuesta cifrada, la descifra y la imprime en pantalla.
* **Registro de actividad (opcional)**: Escribir un log local de las sesiones (fecha/hora, comandos, estado).

***

## Requisitos Técnicos

* **Modularidad**: Separar el código en al menos cuatro componentes (`.cpp`/`.h`):
    * `Comm.cpp` / `Comm.h`: Conexión TCP y encriptación de mensajes.
    * `Crypto.cpp` / `Crypto.h`: Inicialización AES, cifrado/descifrado.
    * `Exec.cpp` / `Exec.h`: Ejecución de comandos y captura de salida.
    * `Persist.cpp` / `Persist.h`: Lógica de alta y baja en el Registro.
    * `main.cpp`: Parseo de parámetros, orquestación y loop principal.

* **Librerías**:
    * `Winsock2`
    * `OpenSSL` o `Crypto++`

* **Manejo de Errores**:
    * Validar la longitud y el formato de la clave AES.
    * Manejar reconexiones en caso de fallo del socket.
    * Comprobar los permisos al escribir en el Registro.

* **Seguridad**:
    * **No hardcodear** credenciales dentro del binario. Permitir, como mínimo, la carga desde un archivo externo si es necesario.

***

## Entregables

* **Repositorio público en GitHub** con:
    * Carpeta `Client/` con el código fuente del RAT.
    * Carpeta `Server/` con un servidor de control simple (puede ser en Python o C++).
    * Ejemplo de archivo de configuración o línea de comandos usada.
    * `README.md` detallando cada módulo, con instrucciones de compilación y ejecución.
    * Log de una sesión de prueba (captura de tráfico o consola).
