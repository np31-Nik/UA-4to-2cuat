/*
	SDS: ejemplo de cifrado sencillo en AES256-CTR mediante streaming y con compresión.

	Compilar con: go build

	Para el cifrado:
	sdscifr -k "contraseña" -i fichentrada -o fichsalida

	Para el descifrado:
	sdscifr -m D -k "contraseña" -i fichentrada -o fichsalida

	**Limitaciones (por sencillez):

	- Derivación de clave.
	Se debería usar una función PBKDF (argon2 por ej.) para derivar una clave a partir de la contraseña en lugar de un hash

	- Cifrado no autentificado.
	No se puede comprobar que el descifrado sea correcto y un atacante puede manipular el texto cifrado sin ser detectado.
	Se debería concatenar el hash de los datos antes de cifrar, emplear un cifrado autentificado (modo GCM, por ej.) o firma digital.
*/
package main

import (
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// check simplifica la gestión de errores en un programa sencillo
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	// hacemos uso del sistema de flags propio de Go para procesar la linea de comandos
	pM := flag.String("m", "C", "modo: C cifrar (por defecto), D descifrar")
	pK := flag.String("k", "", "contraseña (passphrase) para cifrar o descifrar")
	pI := flag.String("i", "", "fichero de entrada (o entrada estándar por defecto)")
	pO := flag.String("o", "", "fichero de salida (o salida estándar por defecto)")
	pC := flag.Bool("c", true, "compresión (activada por defecto)")

	// damos valor a las variables flag
	flag.Parse()

	var key, iv []byte     // clave e iv (slices de 32 bytes -- 256 bits)
	var fin, fout *os.File //ficheros de entrada y salida
	var err error          // contenedor de errores

	iv = make([]byte, 16) // asignar un IV de 128 bits (16 bytes)

	// gestión de la contraseña (passphrase)
	if *pK == "" { // el parámetro de contraseña es obligatorio
		flag.PrintDefaults()
		os.Exit(1)
	} else { // procesar la contraseña mediante sha2 (realmente se utilizaría un PBKDF)
		h := sha256.New()    // creamos el hash
		h.Write([]byte(*pK)) // resumimos la contraseña
		key = h.Sum(nil)     // obtenemos el resumen actual (256 bits)
	}

	if *pI == "" { // si no hay fichero de entrada, se utiliza la entrada estándar
		fin = os.Stdin
	} else { // se abre el fichero pasado en línea de comandos
		fin, err = os.Open(*pI)
		check(err)
		defer fin.Close() // se cierra de forma automática al finalizar el main()
	}

	if *pO == "" { // si no hay fichero de salida, se utiliza la salida estándar
		fout = os.Stdout
	} else { // se crea el fichero pasado en línea de comandos
		fout, err = os.Create(*pO)
		check(err)
		defer fout.Close() // cierre al finalizar al finalizar
	}

	// modo de operación
	var cifrando bool
	if strings.ToUpper(*pM) == "C" { // cifrado
		cifrando = true
		_, err = rand.Read(iv) // crear un iv aleatorio
		check(err)

		_, err = fout.Write(iv) // escribir el iv al inicio del fichero de salida
		check(err)

	} else if strings.ToUpper(*pM) == "D" { // descifrado
		cifrando = false // no cifrando implica descifrando

		_, err = fin.Read(iv) // leer el iv para el descifrado (escrito previamente por el cifrado)
		check(err)

	} else { // modo inválido
		flag.PrintDefaults()
		os.Exit(1)
	}

	t := time.Now() // timestamp inicial para medir el tiempo tardado

	// cifrado/descifrar

	block, err := aes.NewCipher(key) // block es un cifrador en bloque AES con clave key (256 bits)
	check(err)
	S := cipher.NewCTR(block, iv) // S es un cifrador en flujo al aplicar el modo CTR a block con vector inicial iv

	var rd io.Reader      // stream de lectura
	var wr io.WriteCloser // stream de escritura

	if cifrando {
		var enc cipher.StreamWriter // enc es un stream de cifrado que utiliza S y escribe en fout
		enc.S = S
		enc.W = fout

		rd = fin // la lectura se hace desde el fichero de entrada

		if *pC { // compresión activada, se escribe un stream de compresión y éste en el de cifrado (enc)
			wr = zlib.NewWriter(enc)
		} else { // sin compresión, la escritura se realiza en el stream de cifrado
			wr = enc
		}

	} else { // descifrando
		var dec cipher.StreamReader // dec es un stream de descifrado que utiliza S y lee de fin
		dec.S = S
		dec.R = fin

		wr = fout // la escritura se hace directamente en el fichero de salida

		if *pC { // compresión activada, se lee desde un stream de descompresión y éste desde el de descifrado (dec)
			rd, err = zlib.NewReader(dec)
			check(err)
		} else { // sin compresión, la lectura se realiza directamente desde el stream de descifrado
			rd = dec
		}
	}

	_, err = io.Copy(wr, rd) // se copia desde el stream de lectura (rd) al de escritura (wd)
	check(err)
	wr.Close() // se cierra el stream de escritura

	fmt.Println(time.Since(t)) // se imprime el tiempo transcurrido

}
