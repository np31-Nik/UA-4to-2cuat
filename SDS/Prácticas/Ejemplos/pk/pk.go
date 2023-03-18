/*

Este programa demuestra el uso de clave pública (RSA) en una arquitectura cliente servidor sencilla utilizando HTTPS. También demuestra los siguientes conceptos:
- Uso de un sistema genérico de mensajes entre el cliente y servidor
- Cifrado con AES-CTR, hash (SHA256), compresión, encoding (JSON, x509), etc.

Entre otras muchas, algunas limitaciones (por sencillez):
- No se contempla una autoridad certificadora. Si existiera, habría un paso previo en el que la autoridad firmaría las claves públicas de servidor y clientes,
y estos tendrían preinstalada la clave pública de dicha autoridad certificadora
- Sólo se utiliza RSA, no curvas elípticas (aunque podrían ser buena opción pero son más complejas de implementar)
- No hay interactividad, es una mera demostración.


compilación:
go build

arrancar tanto el servidor como el cliente (simultáneo)
sdspk

pd. se pueden reutilizar los certificados incluidos para localhost:
(ver https://letsencrypt.org/docs/certificates-for-localhost/)


*/

package main

import (
	"bytes"
	"compress/zlib"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {

	go server() // servidor en paralelo para poder ejecutar el cliente a continuación
	client()    // el cliente espera a la respuesta del servidor para salir

	// cuando termina el cliente, sale el programa (y todas sus gorutinas)

}

/***
UTILS
***/

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// función para cifrar (AES-CTR 256), adjunta el IV al principio
func encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)    // reservamos espacio para el IV al principio
	rand.Read(out[:16])                 // generamos el IV
	blk, err := aes.NewCipher(key)      // cifrador en bloque (AES), usa key
	chk(err)                            // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)    // ciframos los datos
	return
}

// función para descifrar (AES-CTR 256)
func decrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)     // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)       // cifrador en bloque (AES), usa key
	chk(err)                             // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])     // desciframos (doble cifrado) los datos
	return
}

// función para resumir (SHA256)
func hash(data []byte) []byte {
	h := sha256.New() // creamos un nuevo hash (SHA2-256)
	h.Write(data)     // procesamos los datos
	return h.Sum(nil) // obtenemos el resumen
}

// función para comprimir
func compress(data []byte) []byte {
	var b bytes.Buffer      // b contendrá los datos comprimidos (tamaño variable)
	w := zlib.NewWriter(&b) // escritor que comprime sobre b
	w.Write(data)           // escribimos los datos
	w.Close()               // cerramos el escritor (buffering)
	return b.Bytes()        // devolvemos los datos comprimidos
}

// función para descomprimir
func decompress(data []byte) []byte {
	var b bytes.Buffer // b contendrá los datos descomprimidos

	r, err := zlib.NewReader(bytes.NewReader(data)) // lector descomprime al leer

	chk(err)         // comprobamos el error
	io.Copy(&b, r)   // copiamos del descompresor (r) al buffer (b)
	r.Close()        // cerramos el lector (buffering)
	return b.Bytes() // devolvemos los datos descomprimidos
}

/***
MSG
***/

// mensaje genérico (tanto para peticiones como respuestas)
type msg map[string][]byte // mapa con índice de string y slice de bytes de contenido

// función que hace un post al servidor y devuelve la respuesta
func (m msg) post() (msg, error) {
	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	mBytes, err := json.Marshal(m) // serializamos el mensaje a JSON ([]byte)
	chk(err)

	// hacemos un post pasando un reader al json como body
	r, err := client.Post("https://localhost:10443", "application/octet-stream", bytes.NewReader(mBytes))
	chk(err)

	// extraemos el mensaje de respuesta
	rm := msg{}                            // creamos un mensaje para la respuesta
	rmBytes, err := ioutil.ReadAll(r.Body) // leemos todo el body
	chk(err)
	err = json.Unmarshal(rmBytes, &rm) // deserializamos
	chk(err)

	// comprobamos el status del mensaje de respuesta
	// (protocolo propio, no es el status de http)
	if !strings.EqualFold(string(rm["status"]), "OK") {
		err = errors.New("error en respuesta")
	} else {
		err = nil
	}

	return rm, err // devolvemos el mensaje de respuesta y el error (en su caso)
}

/***
SERVIDOR
***/

// contexto para mantener el estado entre llamadas al handler
var ctxSrv struct {
	privKey   *rsa.PrivateKey // la clave privada del servidor (incluye la pública)
	cliPubKey *rsa.PublicKey  // la clave pública del cliente
}

// gestiona el modo servidor
func server() {
	fmt.Println("Iniciando el servidor...")

	// generamos un par de claves para el servidor (la clave privada incluye la pública)
	var err error
	ctxSrv.privKey, err = rsa.GenerateKey(rand.Reader, 4096) // se puede observar como tarda un poquito en generar
	chk(err)
	ctxSrv.privKey.Precompute() // aceleramos su uso con un precálculo

	http.HandleFunc("/", handler) // asignamos un handler global

	// escuchamos el puerto 10443 con https y comprobamos el error
	chk(http.ListenAndServeTLS("localhost:10443", "localhost.crt", "localhost.key", nil))

}

func handler(w http.ResponseWriter, req *http.Request) {
	// extraemos el mensaje de la petición
	qm := msg{} // creamos un mensaje para la petición

	dec := json.NewDecoder(req.Body) // decoder json para el request
	err := dec.Decode(&qm)           // decodificamos el msg del request
	chk(err)

	rm := msg{} // creamos un mensaje para la respuesta

	switch strings.ToLower(string(qm["cmd"])) { // comprobamos los posibles comandos
	case "xchg":
		// intercambio de claves
		rm["cmd"] = []byte("xchg")                                            // mismo comando
		rm["status"] = []byte("OK")                                           // resultado OK
		rm["srv_pub"] = x509.MarshalPKCS1PublicKey(&ctxSrv.privKey.PublicKey) // clave pública como argumento (necesario x509 para pasar a []byte)

		// extraemos la clave pública del cliente
		ctxSrv.cliPubKey, err = x509.ParsePKCS1PublicKey(qm["cli_pub"]) // necesario x509 para obtener de []byte
		chk(err)

	case "data":
		// intercambio de datos
		txt := "El token enviado es bastante inútil" // texto a enviar
		key := make([]byte, 32)                      // clave aleatoria de cifrado (AES)
		rand.Read(key)
		txtEnc := encrypt(compress([]byte(txt)), key) // comprimimos y ciframos (AES)

		// ciframos la clave de cifrado de datos con la clave pública del destino (cliente)
		keyEnc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, ctxSrv.cliPubKey, key, nil)
		chk(err)

		// firmamos con la clave privada del servidor

		txtDigest := hash([]byte(txt))                                                          // cálculo del resumen
		txtSign, err := rsa.SignPSS(rand.Reader, ctxSrv.privKey, crypto.SHA256, txtDigest, nil) // firmamos el resumen
		chk(err)

		// montamos el mensaje a mandar
		rm["status"] = []byte("OK") // resultado OK
		rm["cmd"] = []byte("data")  // comando datos
		rm["key"] = keyEnc          // la clave de cifrado del mensaje (AES), cifrada con la pública del destino
		rm["data"] = txtEnc         // el mensaje cifrado con AES
		rm["sign"] = txtSign        // la firma del mensaje con la privada del cliente
		rm["digest"] = txtDigest    // hash para comprobar la firma

		//
		// desciframos y comprobamos la firma del cliente
		//

		// desciframos la clave de cifrado con la privada del cliente
		cliKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, ctxSrv.privKey, qm["key"], nil)
		chk(err)
		cliTxt := string(decompress(decrypt(qm["data"], cliKey))) // desciframos el mensaje del cliente
		fmt.Println("SRV::Mensaje del cliente: ", cliTxt)

		// comprobamos la firma
		err = rsa.VerifyPSS(ctxSrv.cliPubKey, crypto.SHA256, qm["digest"], qm["sign"], nil)
		chk(err)
		fmt.Println("SRV::Firma validada")

	default:
		// error
		rm["status"] = []byte("ERROR")
		rm["error"] = []byte("comando inválido")
		rm["request"], err = json.Marshal(qm)
		chk(err)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	enc := json.NewEncoder(w)
	enc.Encode(rm)

}

/***
CLIENTE
***/

// gestiona el modo cliente
func client() {

	// generamos un par de claves (la clave privada incluye la pública)
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 3072) // 3072 bits es significativamente más rápido que 4096
	chk(err)
	rsaPrivKey.Precompute() // aceleramos con un precálculo

	//
	// intercambiamos claves públicas con el servidor
	//
	m := msg{}                                                       // creamos un nuevo mensaje
	m["cmd"] = []byte("xchg")                                        // comando exchange (intercambio)
	m["cli_pub"] = x509.MarshalPKCS1PublicKey(&rsaPrivKey.PublicKey) // clave pública como argumento (necesario x509 para pasar a []byte)

	r, err := m.post() // hacemos el post y obtenemos respuesta
	chk(err)

	srvPubKey, err := x509.ParsePKCS1PublicKey(r["srv_pub"]) // extraemos la clave pública del servidor (necesario x509 para obtener de []byte)
	chk(err)

	//
	// mandamos un mensaje cifrado y firmado
	//
	txt := "El token secreto es 123456" // texto a enviar
	key := make([]byte, 32)             // clave aleatoria de cifrado (AES)
	rand.Read(key)
	txtEnc := encrypt(compress([]byte(txt)), key) // comprimimos y ciframos (AES)

	// ciframos la clave de cifrado de datos con la clave pública del destino (servidor)
	keyEnc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, srvPubKey, key, nil)
	chk(err)

	// firmamos con la clave privada del cliente
	txtDigest := hash([]byte(txt))                                                      // cálculo del resumen
	txtSign, err := rsa.SignPSS(rand.Reader, rsaPrivKey, crypto.SHA256, txtDigest, nil) // firmamos el resumen
	chk(err)

	// montamos el mensaje a mandar
	m = msg{}                 // vaciamos el mensaje anterior
	m["cmd"] = []byte("data") // comando datos
	m["key"] = keyEnc         // la clave de cifrado del mensaje (AES), cifrada con la pública del destino
	m["data"] = txtEnc        // el mensaje cifrado con AES
	m["sign"] = txtSign       // la firma del mensaje con la privada del cliente
	m["digest"] = txtDigest   // hash para comprobar la firma
	r, err = m.post()         // hacemos el post al servidor
	chk(err)

	//
	// desciframos y comprobamos la firma del servidor
	//

	// desciframos la clave de cifrado con la privada del cliente
	srvKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPrivKey, r["key"], nil)
	chk(err)
	srvTxt := string(decompress(decrypt(r["data"], srvKey))) // desciframos el mensaje del servidor
	fmt.Println("CLI::Mensaje del servidor: ", srvTxt)

	// comprobamos la firma
	err = rsa.VerifyPSS(srvPubKey, crypto.SHA256, r["digest"], r["sign"], nil)
	chk(err)
	fmt.Println("CLI::Firma validada")

}
