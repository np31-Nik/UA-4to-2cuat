/*

Este programa muestra cómo enviar un archivo de cliente a servidor
mediante streaming simultáneo sobre https

Recibe el path del archivo de origen como primer parámetro
y utiliza el segundo parámetro como path a escribir en el servidor.

compilación: go build

uso: sdsupl fichorigen fichdestino
(poniendo los paths correctos)

*/

package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	if len(os.Args) != 3 {
		fmt.Println("Uso: sdsupl fichorigen fichdestino")
		os.Exit(1)
	}

	go server() // servidor en paralelo para poder ejecutar el cliente a continuación
	client()
	// cuando termina el cliente, sale el programa (y todas sus gorutinas)
	// el cliente espera a la respuesta del servidor para salir
}

/***
SERVIDOR
***/

// gestiona el modo servidor
func server() {
	fmt.Println("Iniciando el servidor...")
	http.HandleFunc("/", handler) // asignamos un handler global

	// escuchamos el puerto 10443 con https y comprobamos el error
	chk(http.ListenAndServeTLS("localhost:10443", "localhost.crt", "localhost.key", nil))
}

func handler(w http.ResponseWriter, req *http.Request) {
	file, err := os.Create(os.Args[2]) // crea el fichero de destino (servidor)
	chk(err)
	defer file.Close()      // cierra el fichero al salir de ámbito
	t := time.Now()         // timestamp para medir el tiempo
	io.Copy(file, req.Body) // copia desde el Body del request al fichero con streaming

	m := runtime.MemStats{} // obtiene información acerca del uso de memoria
	runtime.ReadMemStats(&m)
	fmt.Println("SRV::", time.Since(t), ":: Fich escrito")  //imprime tiempo
	fmt.Println("SRV:: memoria ", m.TotalAlloc/1024, " KB") // imprime la memoria total

	fmt.Fprintln(w, "¡Fin!") // devuelve un mensaje al cliente
}

/***
CLIENTE
***/

// gestiona el modo cliente
func client() {

	fmt.Println("Iniciando el cliente...")

	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	file, err := os.Open(os.Args[1]) // abrimos el fichero de origen (cliente)
	chk(err)
	defer file.Close() // cerramos al salir de ámbito

	t := time.Now() // timestamp para medir tiempo
	// hacemos un post con formato octet-stream (binario) y ponemos el reader del fichero directamente como Body
	resp, err := client.Post("https://localhost:10443", "application/octet-stream", file)
	chk(err)
	fmt.Println("CLIENTE::", time.Since(t), ":: Post realizado") // imprimimos tiempo
	io.Copy(os.Stdout, resp.Body)                                // copiamos la respuesta a la salida estándar
}
