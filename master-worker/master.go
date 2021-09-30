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
	"strings"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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

func poolGoRutines(chJobs chan com.Job, ip string, puerto string){
	for {
		ruta := ip + ":" + puerto
		job := <- chJobs

		//Conectamos con worker para enviarle los datos
		primos := conectarAWorker(job.Request.Interval, ruta)

		reply := com.Reply{Id: job.Request.Id, Primes: primos}
		fmt.Println("PoolGoRutines envio: " , reply)
		
		encoder := gob.NewEncoder(job.Conn)
		err := encoder.Encode(reply)
		if err != nil {
			job.Conn.Close()
		}
	}
}

func activarWorkerSSH(ip string, puerto string){
	delim := "."
	ip_separada := strings.Split(ip, delim)
	maq := ip_separada[len(ip_separada)-1]
	
	//Sacamos los hosts de esta maquina:
	usr, err := user.Current()
	checkError(err)

	//Buscamos en el fichero de maquinas conocidas
	hostKey, err := knownhosts.New(usr.HomeDir + "/.ssh/known_hosts")
	checkError(err)

	//Buscamos el fichero con la clave: aqui haremos que la clave se llame id_<nom_maq>
	fich := usr.HomeDir + "/.ssh/" + maq
	clave, err := ioutil.ReadFile(fich)
	checkError(err)

	//Ahora agregamos la clave a la conexion
	signer, err2 := ssh.ParsePrivateKey(clave)
	checkErrorPW(err2)

	//Añadimos la config de conexion con ssh
	conf := &ssh.ClientConfig{
		User: usr.Name,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKey,
		Timeout:         0,
	}

	//Generamos el cliente, la conexión ssh se hace por tcp y puerto 22 por defecto
	conexion_ssh := ip + ":22" 
	conn, err := ssh.Dial("tcp", conexion_ssh, conf)
	
	//Comenzamos una sesion y ejecutamos el comando que active el worker
	sesion, err := conn.NewSession()
	checkError(err)
	comando := "/home/a800616/UNI/Tercero/SD/p1-sd-master/master-worker/worker " 
			   + id + " " + puerto
	sesion.Run(comando)
	
	//CUIDAO CON ESTO, porque si cierro conexion entonces igual el worker se va a tomar por culo
	sesion.Close()
	conn.Close()
}

func main() {
	CONN_TYPE := "tcp"
		
	//IMPORTANTE!!!! : Habra que enviar clave publica a dichas maquinas y asegurarnos de que siempre usemos las mismas
	
	//Y tambien tendremos que hacer que se ejecute el worker escuchando a una ip y puerto concreto,
	//habra que pasarlo por parametro al ejecutar con ssh
	
	//Mi idea es leer de un fichero las ip con sus puertos y luego en un for ir llamando a cada
	//gorutine con su ip y ademas lanzar por ssh la ejecucion de los workers
	
	//De momento hardcodeamos el vector de rutas a workers:
	workers := [4]ruta_worker{
		ruta_worker{
			ip: "155.210.154.195",
			puerto: "30000",
		},
		ruta_worker{
			ip: "155.210.154.196",
			puerto: "30001",
		},
		ruta_worker{
			ip: "155.210.154.197",
			puerto: "30002",
		},
		ruta_worker{
			ip: "155.210.154.198",
			puerto: "30003",
		},
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

	//Preparamos la gorutines, que esperaran a recibir algo por el canal
	/*go poolGoRutines(chJobs, IP) 
	go poolGoRutines(chJobs, IP)
	go poolGoRutines(chJobs, IP)
	go poolGoRutines(chJobs, IP)*/
	
	//Activamos todos workers con sus correspodientes ips y puertos a escuchar, tambien
	//arrancamos la gorutines que se conectaran con los workers
	for i := range workers{
		go poolGoRutines(chJobs, workers[i].ip, workers[i].puerto)
		go activarWorkerSSH(workers[i].ip, workers[i].puerto)
	}
	
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

