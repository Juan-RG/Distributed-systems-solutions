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

func poolGoRutines(chRequest chan com.Request, chReply chan com.Reply){
	for {
		dato := <- chRequest
		fmt.Println("PoolGoRutines recibo: " , dato)
		primos := FindPrimes(dato.Interval)
		reply := com.Reply{Id: dato.Id, Primes: primos}
		fmt.Println("PoolGoRutines envio: " , reply)
		chReply <- reply
	}

}

func main() {
	CONN_TYPE := "tcp"
	
	var CONN_PORT, CONN_HOST string
	if len(os.Args) > 1 && os.Args[1] != "" {
		CONN_HOST = os.Args[1]
	} else {
		CONN_HOST = "127.0.0.1"
	}
	
	if len(os.Args) > 2 && os.Args[2] != "" {
		CONN_PORT = os.Args[2]
	} else {
		CONN_PORT = "30000"
	}

	chRequest := make(chan com.Request, 10)
	chReply := make(chan com.Reply, 10)

	go poolGoRutines(chRequest, chReply)
	go poolGoRutines(chRequest, chReply)
	go poolGoRutines(chRequest, chReply)
	go poolGoRutines(chRequest, chReply)


	listener, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
	checkError(err)

	//Establezco todas las conexiones que llegan. El servidor ahora nunca acaba
	for{
		conn, err := listener.Accept()
		//defer conn.Close()
		checkError(err)
		
		go handleClient(conn, chRequest, chReply)
	}

}

func handleClient(conn net.Conn, chRequest chan com.Request, chReply chan com.Reply) {
	//cierro el canal al acabar la funcion. Esto permite el cierre del cliente
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
    decoder := gob.NewDecoder(conn)

    recivoPeticiones := true
    for recivoPeticiones {
	    var request com.Request
	    err := decoder.Decode(&request)
	    if err != nil {
			recivoPeticiones = false
		}
	    //checkError(err)

	    fmt.Println("handleClient recibo: " , request)
	    chRequest <- request

	    reply := com.Reply{-1, nil}
	    for reply.Id != request.Id {
	    	reply = <- chReply
	    	if reply.Id != request.Id {
	    		chReply <- reply
	    	}
	    }
	    fmt.Println("handleClient recibo: " , reply)
	    fmt.Println("Consecuencia " , reply.Id, " --- ", request.Id)
	    //listaPrimos := FindPrimes(request.Interval)

		//quitar el id hardcode												//Todo: Poner el id o incremental o aleatorio -> Mas facil a mi parecer aleatorio
		//reply := com.Reply{Id: request.Id, Primes: listaPrimos}

		err = encoder.Encode(reply)
		if err != nil {
			recivoPeticiones = false
		}
    }
}

