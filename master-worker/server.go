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
/*
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
}*/

//Aqui mediante un puerto y una ip enviaremos el struct de primos con gob
func conectarAWorker(intervalo com.TPInterval, ip string) []int {
//func conectarAWorker(job Job, ip string) []int {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	checkError(err)

	worker, _ := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	
	defer worker.Close() //Igual hay que cerrar al final de la funcion y no al final del prog
	
	//Enviamos el intervalo para que lo procese el worker
	err = gob.NewEncoder(worker).Encode(intervalo)
	
	if err != nil {
		fmt.Println(err)
	}
		
	//Ahora habra que escuchar la ip y puerto para recibir los primos
	var primos []int
	err = gob.NewDecoder(worker).Decode(&primos)
	checkError(err)

	return primos
}

func poolGoRutines(chJobs chan com.Job, ip string){
	for {
		job := <- chJobs

		//Conectamos con worker para enviarle los datos
		primos := conectarAWorker(job.Request.Interval, ip)

		reply := com.Reply{Id: job.Request.Id, Primes: primos}
		fmt.Println("PoolGoRutines envio: " , reply)
		
		encoder := gob.NewEncoder(job.Conn)
		err := encoder.Encode(reply)
		if err != nil {
			job.Conn.Close()
		}
	}
}

func main() {
	CONN_TYPE := "tcp"
	
	//Recibiremos por fichero unos workers y hay que activarlos mediente ssh (nos conectamos a central y ejecutamos comando para activarlos???)
	//Luego habra que tener un fichero que se llame worker y ejecute el findPrimes?
	IP := "155.210.154.196:30000" //Tendremos un fichero con ip y puerto?

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

	//Preparamos la gorutines, que esperaran a recibir algo por el canal
	//Habra que meter otro parametro con la maquina a la que enviara la gorutine?
	go poolGoRutines(chJobs, IP) 
	/*go poolGoRutines(chJobs, IP)
	go poolGoRutines(chJobs, IP)
	go poolGoRutines(chJobs, IP)*/


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

	//Recibimos los datos del cliente
    decoder := gob.NewDecoder(conn)

    reciboPeticiones := true
	
    for reciboPeticiones {
	    var request com.Request
		//Transformamos lo bytes que nos llegan al struct 
	    err := decoder.Decode(&request)
	    if err != nil {
			reciboPeticiones = false
			conn.Close()
			fmt.Println("Okey ")
			break
		}
	    //checkError(err)
	    fmt.Println("handleClient recibo: " , request)
		
		//Creamos el trabajo (Conexion y datos a procesar del cliente)
	    job := com.Job{conn, request}
		//Enviamos al canal de las gorutines para que procesen los datos
	    chJobs <- job
    }
}

