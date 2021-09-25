/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes al trabajo 1
*/
package main

import (
	"fmt"
	"net"
	"os"
	"com"
    "bytes"
    "compress/gzip"
    "encoding/gob"
    "io/ioutil"
    "log"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func IsPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
// 		intervalo [interval.A, interval.B]
func FindPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.A; i <= interval.B; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

//Pre []byte array
//post return a Request
func Decode(data []byte) com.Request {

	r := com.Request{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&r)
	if err != nil {
		log.Fatal(err)
	}
	return r
}
//function for descompress the data
func Decompress(s []byte) []byte {

	rdr, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Fatal(err)
	}
	rdr.Close()
	//descomprime un array de bytes
	return data
}

//funcion para  serializar
func EncodeToBytes(p interface{}) []byte {

    buf := bytes.Buffer{}
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(p)
    if err != nil {
        log.Fatal(err)
    }
	//Retorno un array de bytes de un struct
    return buf.Bytes()
}

//Funcion para comprimir
func Compress(s []byte) []byte {

    zipbuf := bytes.Buffer{}
    zipped := gzip.NewWriter(&zipbuf)
    zipped.Write(s)
    zipped.Close()
	//Comprimo un array de bites
    return zipbuf.Bytes()
}

func main() {
	CONN_TYPE := "tcp"
	CONN_HOST := "127.0.0.1"
	CONN_PORT := "30000"

	listener, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
	checkError(err)
	//Establezco todas las conexiones que llegan. El servidor ahora nunca acaba
	for{
		conn, err := listener.Accept()
		//defer conn.Close()
		checkError(err)
		
		go handleClient(conn)
	}

}

func handleClient(conn net.Conn) {
	//cierro el canal al acabar la funcion. Esto permite el cierre del cliente
	defer conn.Close()

	//creo un buffer de 1024. Por protocolo TCP. Llega menos datos por que es un struct formado por 3 ints
  	data := make([]byte, 1024)
  	//leo el struct
  	n, err := conn.Read(data)

  	fmt.Println(data[0:n])

  	//Descomprimo los datos recibidos de 0 -> N(numero de bytes recibidos)
	dataIn := Decompress(data[0:n])
	//lo decodifico para obtener la respuesta
	request := Decode(dataIn)

	//Saco los primos del intervalo
	listaPrimos := FindPrimes(request.Interval)

	//quitar el id hardcode												//Todo: Poner el id o incremental o aleatorio -> Mas facil a mi parecer aleatorio
	reply := com.Reply{Id: 0, Primes: listaPrimos}

	//Serializo la respuesta
	dataOut := EncodeToBytes(reply)
    //Lo comprimo para enviar menos bytes
    dataOut = Compress(dataOut)
    //envio la lista de primos.
    conn.Write(dataOut)

	fmt.Println(listaPrimos, err)


}

