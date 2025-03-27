# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```


## Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).


### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `


### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).

## Resolución

### Ejercicio 1

Se define un script de bash llamado `generar-compose.sh` para la creación de un archivo Docker Compose dinámico basado en la cantidad de clientes especificada. Para ejecutar este script, se debe utilizar la siguiente línea de comando:

```
./generar-compose.sh <Nombre del archivo> <N° Clientes>
```

Se recomienda utilizar el nombre `docker-compose-dev.yaml` como archivo de salida.

### Implementación

El script se implementó completamente en bash, siguiendo los pasos detallados a continuación:

1. **Parámetros de entrada**: 
    - `$1`: Nombre del archivo de salida.
    - `$2`: Cantidad de clientes.

2. **Creación del archivo**: 
    - Se utiliza el comando `touch` para crear el archivo especificado en `$1`.

3. **Escritura de configuración inicial**:
    - Se escribe la configuración base del archivo Docker Compose, incluyendo la definición del servidor.

4. **Generación dinámica de clientes**:
    - Se utiliza un bucle `for` que itera desde `1` hasta el número de clientes especificado en `$2`.
    - En cada iteración, se agrega un bloque de configuración para un cliente, incluyendo una variable de entorno que define su `ID`.

5. **Configuración de red**:
    - Al finalizar, se agrega la configuración de la red que será utilizada por los contenedores.


### Ejercicio 2

Se solicita modificar tanto el cliente como el servidor para evita  la necesidad de reconstruir las imágenes de Docker en caso de que se realicen cambios en los archivos de configuración:
- config.yaml para el cliente
- config.ini para el servidor

Los volúmenes son esenciales para separar los datos de la lógica de la aplicación, facilitando el desarrollo y la administración de contenedores. Los volúmenes permiten persistir datos y compartir archivos entre el host y los contenedores. Como son unidades que se montan luego, si los cambio, no es necesario generar nuevamente una imagen.

Para que esto se cumpla también se eliminó de los Dockerfile de cada servicio (servidor y cliente) la copia de estos archivos de configuración.

De esta forma, al script de `generar-compose.sh` se agrega:
- Servidor: 
```
volumes: 
      - ./server/config.ini:/config.ini
```

- Cliente: 
```
    volumes:
      - ./client/config.yaml:/config.yaml
```

La ejecución es igual al ejercicio 1.

### Ejercicio 3

El script validar-echo-server.sh verifica el correcto funcionamiento del servidor utilizando netcat dentro de un contenedor Docker. Dado que el servidor es un echo server, el script envía un mensaje y espera recibir el mismo mensaje como respuesta.


El script, nuevamente, se implementó completamente en bash, siguiendo los pasos detallados a continuación:

1) Se definen las siguientes variables:
- `NETWORK`: nombre de la red que se va a usar para la comunicacion
- `PORT`: puerto en el que esta escuchando el servidor
- `MESSAGE`: mensaje de prueba para enviar al servidor

2) Se ejecuta `netcat` dentro de un contenedor docker con una imagen `busybox` para poder ejecutar el comando `nc`. Una vez enviado el mensaje al servidor se almacena la respuesta en la variable `RESPONSE`

3) Por ultimo se verifica que la respuesta coincida con el mensaje enviado imprimiento `success` o `fail` si no hay coincidencia.

La ejecución de este script es similar a la del ejercicio 1, salvo que no recibe parametros:
```
./validar-echo-server.sh
```

### Ejercicio 4

En este ejercicio se pide que ambos serivicios, tanto el cliente como el servidor terminen de forma graceful al recibir la signal SIGTERM. 

#### Cliente
Para el caso del cliente se sigue los siguientes paso dentro del código:

1) Se captura o se hace un catch de la señal SIGTERM. Esto se realiza mediante un channel `sigChan`. Dentro de este channel usamos `signal.Notify` para regsitrar la señal `SIGTERM` y enviarla al canal

2) El manejo de esta señal se hace mediante una goroutine que es una función o método que se ejecuta de forma concurrente. 

3) Cuando llega la señal se imprime un mensaje informativo y si existe una conexión abierta (por que puede que no la haya) se cierra correctamente y se llama a `os.Exit(0)` para finalizar el programa de manera controlada

#### Servidor
Para el caso del servidor se sigue los siguientes paso dentro del código:

1) Para capturar la señal de `SIGTERM` se usa signal que es un paquete de python que permite manejar señales del sistema. Cuando el proceso recibe esta señal ejecuta la función signal_handler

2) Esta función cierra, si es que hay, la conexión activa con el cliente y cierra además el socket del servidor

3) Por último imprime estos pasos y llama a `exit(0)` para terminar el proceso de forma controlada

### Ejercicio 5

A partir de este ejercicio, se inicia la Parte 2 del trabajo, enfocada en la comunicación entre servicios. En este caso, se modelan agencias de lotería (clientes) que interactúan con la central de lotería nacional (servidor).

#### Implementación
Para la implementación de las distintas apuestas se creo una estructura que contenga los datos de la misma y que tenga métodos encargados de validar los campos y de serializarlos para su correcto envio.

La agencia (cliente) lo que hace es crear una nueva apuesta obteniendo los valores de variables de entornos seteadas en el docker compose mediante el script `generar-compose.sh`

Una vez que se envia el cliente la apuesta pasa a la espera de un ACK de la central de lotería indicando que la apuesta se realizó con exito.

Con respecto a la central (servidor), una vez establecida la conexión pasa a esperar que la agencia le envie una apuesta. Cuando recibe el mensaje de una apuesta comienza a leer de forma segura evitando short reads. Una vez que finaliza, crea una Bet mediante la clase proporcionada por la cátedra y la almacena. Luego de almacenarla envia un ACK a la agencia confirmando el almacenamiento.

#### Protocolo
Dentro de la serialización se utilizó el siguiente protocolo:

- 1 Byte inicial para identificar el tipo de mensaje. Dentro de los tipos de mensajes tenemos:
  - `0x01` se envia una apuesta
  - `0x11` ACK

- El resto de los campos se serializan de la siguiente forma:
  - 1 Byte para describir la longitud del campo a mandar
  - (1...255) Bytes para el valor del campo

Este diseño permite manejar mensajes de hasta 255 caracteres por campo, lo cual es más que suficiente para los datos de la apuesta.

### Ejercicio 6

Para el ejercicio 6 se pide modificar la agencia (cliente) para que envie varias apuestas a la vez. Estas apuestas ahora se simulan mediante un archivo `.csv` y no mediante variables de entorno como en el ejercicio pasado. Estos archivos se insertaran como volumenes de la misma forma que se mostró en el ejercicio 2. 

Para la correcta lectura de las apuestas se optó por crear una nueva estructura llamada Reader que servirá para proporcionarme los distintos chunks de data para enviar. 

Por cada chunk de data que enviemos, la agencia (cliente) esperará la confirmación (ACK) de la central de lotería para continuar enviando chunks. 

#### Protocolo
La implementación del protocolo cambia un poco a partir de ahora siguiendo estos criterios

- 1 Byte inicial para identificar el tipo de mensaje. Dentro de los tipos de mensajes tenemos:
  - `0xNN` para identificar la cantidad de apuestas que se están enviando dentro del chunk.
  - `0xFF` ACK
  - `0x00` para identificar que ya se enviaron todas las apuestas.

Como se pide enviar chunks que no superen los 8kB se busco la forma de maximizar este numero. Es por eso que se decide cambiar el criterio de las longitudes de los campos siguiendo ahora estos máximos:

- nombre: máximo 30 bytes
- apellido: máximo 15 bytes
- documento: máximo 8 bytes
- nacimiento: máximo 10 bytes
- numero: máximo 10 bytes

Si alguno de estos criterios no se cumple, la apuesta no podrá ser creada.  

- El resto de los campos se serializan de la siguiente forma:
  - 1 Byte para describir la longitud del campo a mandar
  - (1...N) Bytes para el valor del campo

Teniendo esto en cuenta, en el peor de los casos, una apuesta tendrá un tamaño de 79 bytes que nos da un máximo de 101 apuestas por chunk (tomamos 100 como máximo)

Con respecto al servidor, lo que hará será loopear hasta que la cantidad de apuestas que se le envíen sea 0. Dentro del loop, va a procesar el chunk de apuestas leyendo por cada apuesta la longitud de los campos y creando una apuesta y almacenandola. Por cada chunk que se almacena correctamente se devuelve un ACK a modo de confirmación.

#### Validación y Manejo de Errores

Para garantizar la robustez del sistema, tanto el cliente como el servidor implementan mecanismos de validación y manejo de errores:

- **Cliente**:
  - Antes de enviar un chunk, valida que todas las apuestas cumplan con los criterios de longitud establecidos. Si alguna apuesta no es válida, se descarta y se registra un mensaje de error en los logs.

- **Servidor**:
  - Valida cada apuesta recibida en el chunk antes de almacenarla. Si alguna apuesta no es válida, se descarta y se registra un mensaje de error en los logs.
  - Si ocurre un error al procesar un chunk completo, el servidor responde con un código de error específico y no almacena ninguna de las apuestas del chunk. Esto asegura la consistencia de los datos almacenados.

Estos mecanismos permiten manejar errores de forma controlada y aseguran que el sistema sea resiliente frente a fallos en la comunicación o datos mal formateados.

### Ejercicio 7

La complejidad que se agrega en esta parte del ejercicio consiste en que la agencia una vez enviado todas las apuestas, envia un mensaje pidiendo a sus ganadores. 

Para realizar esto se suma una nueva acción al protocolo y es que el primer byte que identifica los distintos tipos de mensajes ahora va a poder enviar:
- `0xF0` para solicitar la cantidad de ganadores de su agencia.

### Cambios en el servidor para el ejercicio 7

En el ejercicio 7, se introduce la funcionalidad para que las agencias puedan solicitar la lista de ganadores una vez que hayan enviado todas sus apuestas. Esto implica los siguientes cambios en el servidor:

1. **Recepción de la solicitud de ganadores**:
   - El servidor ahora debe esperar un mensaje con el identificador `0xF0` después de que todas las apuestas han sido enviadas por una agencia.
   - Cuando recibe este mensaje, el servidor tiene que esperar que todas las agencias hayan enviados sus apuestas. Debido a esto se guarda el socket de esa agencia y procede a conectarse a otros clientes almacenando las apuestas


2. **Sincronización de agencias**:
    - Para saber cuantas agencias el serivdor tiene que esperar se utiliza una variable de entorno previamente seteada por el script con el que se genera el docker compose.
    - Una vez enviada todas las apuestas procede a obtener los ganadores de cada agencia.

3. **Envío de ganadores**:
   - El servidor utiliza envia la cantidad de ganadores a cada agencia una vez que las agencias se lo hayan solicitado

#### Flujo actualizado del servidor

1. **Recepción de apuestas**:
   - Igual que en el ejercicio 6, el servidor recibe las apuestas en chunks y almacena las válidas.

2. **Notificación de finalización**:
   - Una vez que una agencia termina de enviar sus apuestas, envía el mensaje `0xF0` para notificar que ha terminado y solicita los ganadores.

3. **Sorteo**:
   - El servidor espera a que las 5 agencias envíen el mensaje `0xF0`. Una vez que todas las agencias han notificado, el servidor realiza el sorteo utilizando las funciones `load_bets` y `has_won`.

4. **Respuesta a las agencias**:
   - El servidor responde a cada agencia con la lista de ganadores correspondiente a esa agencia.
