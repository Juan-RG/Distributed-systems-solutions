package main

import (
	"fmt"
	"net"
	"com"
    "encoding/gob"
    "os"

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


func main() {

	//Necesitamos que pase por parametro el puerto y la ip a la que tenemos que escuchar
	CONN_TYPE := "tcp"
	CONN_HOST := "155.210.154.196"
	CONN_PORT := "30000"

	listener, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
	checkError(err)
	//Establezco todas las conexiones que llegan. El servidor ahora nunca acaba
	for{
		conn, err := listener.Accept()
		checkError(err)
		
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	//cierro el canal al acabar la funcion. Esto permite el cierre del cliente
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
    decoder := gob.NewDecoder(conn)

    var intervalo com.TPInterval
	err := decoder.Decode(&intervalo)
	checkError(err)

	//Buscamos los primos y enviamos directamente
	listaPrimos := FindPrimes(intervalo)
	err = encoder.Encode(listaPrimos)
	checkError(err)
}