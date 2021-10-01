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
	"utils"
)



func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

//Aqui mediante un puerto y una ip enviaremos el struct de primos con gob
func conectarAWorker(intervalo com.TPInterval, ip string) []int {
//func conectarAWorker(job Job, ip string) []int {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	checkError(err)

	worker, _ := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	
	defer worker.Close() //Igual hay que cerrar al final de la funcion y no al final del prog
	
	//Enviamos el intervalo para que lo procese el worker
	fmt.Println("Envio al worker ", ip)
	err = gob.NewEncoder(worker).Encode(intervalo)
	checkError(err)
		
	//Ahora habra que escuchar la ip y puerto para recibir los primos
	var primos []int
	err = gob.NewDecoder(worker).Decode(&primos)
	checkError(err)

	return primos
}

func poolGoRutines(chJobs chan com.Job, ip string, puerto string){
	ruta := ip + ":" + puerto
	for {

		job := <- chJobs
		fmt.Println("He leido del canal: ", job)
		
		fmt.Println("Voy a enviar a ", ruta)
		//Conectamos con worker para enviarle los datos
		primos := conectarAWorker(job.Request.Interval, ruta)

		reply := com.Reply{Id: job.Request.Id, Primes: primos}
		fmt.Println("PoolGoRutines envio: " , reply)
		
		encoder := gob.NewEncoder(job.Conn)
		err := encoder.Encode(reply)
		defer job.Conn.Close()
		fmt.Println(err)
/*		if err != nil {
			job.Conn.Close()
		}*/
	}
}

func activarWorkerSSH(ip string, puerto string){
	//fmt.Println("Entramos en activarSSH")
	
	ssh, err := utils.NewSshClient(
		"juan",																				//ToDo: Poner como argumento
		ip,
		22,
		"C:/Users/Juan/.ssh/id_rsa",															//ToDo: Poner como argumento
		"")
	if err != nil {
		fmt.Printf("SSH init error %v", err)
	} else {
		comando := "/home/juan/Escritorio/SD/worker " + ip + " " + puerto+ "&"
		//comando := "/home/a800616/UNI/Tercero/SD/p1-sd-master/master-worker/worker"

		//output, err := ssh.RunCommand(comando)
		ssh.RunCommand(comando)
		fmt.Println("comando lanzado")
	}
}

func main() {
	CONN_TYPE := "tcp"
	

	
	//De momento hardcodeamos el vector de rutas a workers:
	workers := []com.Ruta_worker{
		com.Ruta_worker{
			Ip: "192.168.1.228",
			Puerto: "40000",
		},
/*		com.Ruta_worker{
			Ip: "155.210.154.196",
			Puerto: "40000",
		},
		com.Ruta_worker{
			Ip: "155.210.154.193",
			Puerto: "40000",
		},
		com.Ruta_worker{
			Ip: "155.210.154.198",
			Puerto: "40000",
		},
*/
	}

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

	//Activamos todos workers con sus correspodientes ips y puertos a escuchar, tambien
	//arrancamos la gorutines que se conectaran con los workers
	for i := range workers{
		//Activamos los workers
		activarWorkerSSH(workers[i].Ip, workers[i].Puerto)
		go poolGoRutines(chJobs, workers[i].Ip, workers[i].Puerto)
	}

	fmt.Println("Salgo")
	listener, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
	checkError(err)

	//Establezco todas las conexiones que llegan. El servidor ahora nunca acaba
	for{
		conn, err := listener.Accept()
		//defer conn.Close()
		checkError(err)

		go handleClient(conn, chJobs)
	}

	//Comando para matar workers
	//kill -9 $(ps aux -u juan | grep "/home/juan/Escritorio/SD/worker 192.168.1.228 40000" | head -1 | tr -s ' ' | cut -d " " -f 2)


}

func handleClient(conn net.Conn, chJobs chan com.Job) {
	//Recibimos los datos del cliente
    decoder := gob.NewDecoder(conn)

  //  reciboPeticiones := true
	
   // for reciboPeticiones {
	    var request com.Request
		//Transformamos lo bytes que nos llegan al struct 
	    err := decoder.Decode(&request)
	//    if err != nil {
	//		reciboPeticiones = false
	//		conn.Close()
	//		break
	//	}
	    //checkError(err)
	    fmt.Println(err,"handleClient recibo: " , request)
		
		//Creamos el trabajo (Conexion y datos a procesar del cliente)
	    job := com.Job{conn, request}
		fmt.Println("He creado el job")
		//Enviamos al canal de las gorutines para que procesen los datos
	    chJobs <- job
		fmt.Println("He enviado el job")
    //}
}

