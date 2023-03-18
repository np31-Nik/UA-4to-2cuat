/*

Este programa demuestra una arquitectura cliente servidor sencilla utilizando sockets TCP y TLS
El cliente envía líneas desde la entrada estandar y el servidor le devuelve un reconomiento de llegada (acknowledge).
El servidor es concurrente, siendo capaz de manejar múltiples clientes simultáneamente.
Las entradas se procesan mediante un scanner (bufio).

compilación:
go build

arrancar el servidor:
sdstls srv

arrancar el cliente:
sdstls cli

pd. Comando openssl para generar el par certificado/clave para localhost:
(ver https://letsencrypt.org/docs/certificates-for-localhost/)

openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

*/

package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"os"
)

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	fmt.Println("sdstls :: un ejemplo sobre la gestión de conexiones TCP con TLS en Go.")
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "srv":
			fmt.Println("Entrando en modo servidor...")
			server()
		case "cli":
			fmt.Println("Entrando en modo cliente...")
			client()
		default:
			fmt.Println("Parámetro '", os.Args[1], "' desconocido. Introduce srv para funcionalidad de servidor y cli para funcionalidad de cliente")
		}
	} else {
		fmt.Println("Introduce srv para funcionalidad de servidor y cli para funcionalidad de cliente")
	}
}

// gestiona el modo servidor
func server() {
	// cargamos el par certificado / clave privada
	cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
	chk(err)

	// asignamos dicho par a la configuración de TLS
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	// creamos un listener para escuchar el puerto 1337
	ln, err := tls.Listen("tcp", "127.0.0.1:1337", cfg)
	chk(err)
	defer ln.Close() // nos aseguramos que cerramos las conexiones aunque el programa falle

	for { // búcle infinito, se sale con ctrl+c o matando el proceso
		conn, err := ln.Accept() // para cada nueva petición de conexión
		chk(err)

		go func() { // lanzamos un cierre (lambda, función anónima) en concurrencia

			_, port, err := net.SplitHostPort(conn.RemoteAddr().String()) // obtenemos el puerto remoto para identificar al cliente (decorativo)
			chk(err)

			fmt.Println("conexión: ", conn.LocalAddr(), " <--> ", conn.RemoteAddr())

			scanner := bufio.NewScanner(conn) // el scanner nos permite trabajar con la entrada línea a línea (por defecto)

			for scanner.Scan() { // escaneamos la conexión
				fmt.Println("cliente[", port, "]: ", scanner.Text()) // mostramos el mensaje del cliente
				fmt.Fprintln(conn, "ack: ", scanner.Text())          // enviamos ack al cliente
			}

			conn.Close() // cerramos al finalizar el cliente (EOF se envía con ctrl+d o ctrl+z según el sistema)
			fmt.Println("cierre[", port, "]")
		}()
	}
}

func client() {
	// desactivamos la comprobación del certificado (útil en desarrollo con certificado autofirmado)
	cfg := &tls.Config{InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", "127.0.0.1:1337", cfg) // llamamos al servidor con esa configuración
	chk(err)
	defer conn.Close() // es importante cerrar la conexión al finalizar

	fmt.Println("conectado a ", conn.RemoteAddr())

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	for keyscan.Scan() { // escaneamos la entrada
		fmt.Fprintln(conn, keyscan.Text())         // enviamos la entrada al servidor
		netscan.Scan()                             // escaneamos la conexión (se bloquea hasta recibir información)
		fmt.Println("servidor: " + netscan.Text()) // mostramos mensaje desde el servidor
	}

}
