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

func poolGoRutines(chJobs chan com.Job){
	for {
		job := <- chJobs
		encoder := gob.NewEncoder(job.Conn)
		primos := FindPrimes(job.Request.Interval)
		reply := com.Reply{Id: job.Request.Id, Primes: primos}
		err := encoder.Encode(reply)
		if err != nil {
			job.Conn.Close()
		}
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

	chJobs := make(chan com.Job, 10)

	go poolGoRutines(chJobs)
	go poolGoRutines(chJobs)
	go poolGoRutines(chJobs)
	go poolGoRutines(chJobs)
	go poolGoRutines(chJobs)
	go poolGoRutines(chJobs)

	listener, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
	checkError(err)

	//Establezco todas las conexiones que llegan. El servidor ahora nunca acaba

	for{
		conn, err := listener.Accept()
		//defer conn.Close()
		checkError(err)
		
		go handleClient(conn, chJobs)

	}

}

func handleClient(conn net.Conn, chJobs chan com.Job) {
	//cierro el canal al acabar la funcion. Esto permite el cierre del cliente
	//defer conn.Close()

	
    decoder := gob.NewDecoder(conn)
    var request com.Request
    err := decoder.Decode(&request)
	checkError(err)
    
    job := com.Job{conn, request}
    chJobs <- job
    
}

