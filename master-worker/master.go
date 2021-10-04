/*
* AUTOR: Miguel Beltrán y Juan Antonio Rodríguez
*
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: master.go
* DESCRIPCIÓN: master con 6 gorutines para lanzar 6 peticiones por servidor
*/
package main

import (
	"fmt"
	"net"
	"com"
    "encoding/gob"
    "os"
	"utils"
	"io/ioutil"
	"encoding/json"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

//Aqui mediante un puerto y una ip enviaremos el struct de primos con gob
func conectarAWorker(intervalo com.TPInterval, ip string) []int {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	checkError(err)

	worker, _ := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)
	
	defer worker.Close() //Igual hay que cerrar al final de la funcion y no al final del prog
	
	//Enviamos el intervalo para que lo procese el worker
	err = gob.NewEncoder(worker).Encode(intervalo)
	checkError(err)
		
	//Ahora habra que escuchar la ip y puerto para recibir los primos
	var primos []int
	err = gob.NewDecoder(worker).Decode(&primos)
	checkError(err)

	return primos
}

func poolGoRutines(chJobs chan com.Job, ip string, puerto string){

	for {
		job := <- chJobs
		//Conectamos con worker para enviarle los datos
		primos := conectarAWorker(job.Request.Interval, ip + ":" + puerto)
		reply := com.Reply{Id: job.Request.Id, Primes: primos}
		encoder := gob.NewEncoder(job.Conn)
		encoder.Encode(reply)
		defer job.Conn.Close()
	}

}


func activarWorkerSSH(ip string, puerto string){	
	ssh, err := utils.NewSshClient(
		"a805001",																				//ToDo: Poner como argumento
		ip,
		22,
		"/home/a805001/.ssh/id_rsa",														//ToDo: Poner como argumento
		"")
	if err != nil {
		fmt.Printf("SSH init error %v", err)
	} else {
		comando := "/home/a805001/Desktop/SD/master-worker/worker " + ip + " " + puerto+ "&"				//ToDo: poner como argumento
		//comando := "/home/a800616/UNI/Tercero/SD/p1-sd-master/master-worker/worker"
		fmt.Println("Ejecuto comando")
		ssh.RunCommand(comando)
		fmt.Println("salgo")
	}
}


type Rutas struct {
    Workers []Ruta_worker `json:"server"`
}


type Ruta_worker struct {
    Ip   string `json:"ip"`
    Puerto   string `json:"puerto"`
}





//----------------------

func main() {
	CONN_TYPE := "tcp"
	
 	
 // Open our jsonFile
    jsonFile, err := os.Open("workers.json")
    // if we os.Open returns an error then handle it
    if err != nil {
        fmt.Println(err)
    }
    // defer the closing of our jsonFile so that we can parse it later on
    defer jsonFile.Close()

    // read our opened xmlFile as a byte array.
    byteValue, _ := ioutil.ReadAll(jsonFile)

    // we initialize our Users array
    var rutas Rutas

    // we unmarshal our byteArray which contains our
    // jsonFile's content into 'users' which we defined above
    json.Unmarshal(byteValue, &rutas)


	var CONN_PORT, CONN_HOST string
	if len(os.Args) > 1 && os.Args[1] != "" {
		CONN_HOST = os.Args[1]
	} else {
		CONN_HOST = "127.0.0.1"
	}
	
	if len(os.Args) > 2 && os.Args[2] != "" {
		CONN_PORT = os.Args[2]
	} else {
		CONN_PORT = "40000"
	}

	chJobs := make(chan com.Job, 10)

	//Activamos todos workers con sus correspodientes ips y puertos a escuchar, tambien
	//arrancamos la gorutines que se conectaran con los workers
	for i := range rutas.Workers{
		fmt.Println(i, rutas.Workers[i])
		//Activamos los workers
		activarWorkerSSH(rutas.Workers[i].Ip, rutas.Workers[i].Puerto)
		go poolGoRutines(chJobs, rutas.Workers[i].Ip, rutas.Workers[i].Puerto)
		go poolGoRutines(chJobs, rutas.Workers[i].Ip, rutas.Workers[i].Puerto)
		go poolGoRutines(chJobs, rutas.Workers[i].Ip, rutas.Workers[i].Puerto)
		go poolGoRutines(chJobs, rutas.Workers[i].Ip, rutas.Workers[i].Puerto)
		go poolGoRutines(chJobs, rutas.Workers[i].Ip, rutas.Workers[i].Puerto)
		go poolGoRutines(chJobs, rutas.Workers[i].Ip, rutas.Workers[i].Puerto)

	}

	listener, err := net.Listen(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
	checkError(err)

	//Establezco todas las conexiones que llegan. El servidor ahora nunca acaba
	for {
		conn, err := listener.Accept()
		//defer conn.Close()
		checkError(err)

		go handleClient(conn, chJobs)
	}

	//Comando para matar workers
	//kill -9 $(ps aux -u juan | grep "/home/juan/Escritorio/SD/worker 192.168.1.228 40000" | head -1 | tr -s ' ' | cut -d " " -f 2)
	//kill -9 $(ps -u a805001 | grep "worker" | head -1 | tr -s ' ' | cut -d " " -f 2)
}

func handleClient(conn net.Conn, chJobs chan com.Job) {
	
	//Recibimos los datos del cliente
    decoder := gob.NewDecoder(conn)

    var request com.Request
	//Transformamos lo bytes que nos llegan al struct 
    decoder.Decode(&request)

	//Creamos el trabajo (Conexion y datos a procesar del cliente)
    job := com.Job{conn, request}

	//Enviamos al canal de las gorutines para que procesen los datos
    chJobs <- job
}

