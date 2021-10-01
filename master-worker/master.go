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
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type SshClient struct {
	Config *ssh.ClientConfig
	Server string
}

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
	for {
		ruta := ip + ":" + puerto
		job := <- chJobs
		fmt.Println("He leido del canal: ", job)
		
		fmt.Println("Voy a enviar a ", ruta)
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

func NewSshClient(user string, host string, port int, privateKeyPath string, privateKeyPassword string) (*SshClient, error) {
	// read private key file
	pemBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Reading private key file failed %v", err)
	}
	// create signer
	signer, err := signerFromPem(pemBytes, []byte(privateKeyPassword))
	if err != nil {
		return nil, err
	}
	// build SSH client config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}

	client := &SshClient{
		Config: config,
		Server: fmt.Sprintf("%v:%v", host, port),
	}

	return client, nil
}

// Opens a new SSH connection and runs the specified command
// Returns the combined output of stdout and stderr
func (s *SshClient) RunCommand(cmd string) {
	// open connection
	conn, err := ssh.Dial("tcp", s.Server, s.Config)
	if err != nil {
		fmt.Printf("Dial to %v failed %v", s.Server, err)
	}
	defer conn.Close()

	// open session
	session, err := conn.NewSession()
	if err != nil {
		fmt.Printf("Create session for %v failed %v", s.Server, err)
	}
	defer session.Close()
	fmt.Println(cmd)
	session.Run(cmd)

	// run command and capture stdout/stderr
	//output, err := session.CombinedOutput(cmd)

	//return fmt.Sprintf("%s", output), err
}

func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {

	// read pem block
	err := errors.New("Pem decode failed, no key found")
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, err
	}

	// handle encrypted key
	if x509.IsEncryptedPEMBlock(pemBlock) {
		// decrypt PEM
		pemBlock.Bytes, err = x509.DecryptPEMBlock(pemBlock, []byte(password))
		if err != nil {
			return nil, fmt.Errorf("Decrypting PEM block failed %v", err)
		}

		// get RSA, EC or DSA key
		key, err := parsePemBlock(pemBlock)
		if err != nil {
			return nil, err
		}

		// generate signer instance from key
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, fmt.Errorf("Creating signer from encrypted key failed %v", err)
		}

		return signer, nil
	} else {
		// generate signer instance from plain key
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing plain private key failed %v", err)
		}

		return signer, nil
	}
}

func parsePemBlock(block *pem.Block) (interface{}, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing PKCS private key failed %v", err)
		} else {
			return key, nil
		}
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing EC private key failed %v", err)
		} else {
			return key, nil
		}
	case "DSA PRIVATE KEY":
		key, err := ssh.ParseDSAPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing DSA private key failed %v", err)
		} else {
			return key, nil
		}
	default:
		return nil, fmt.Errorf("Parsing private key failed, unsupported key type %q", block.Type)
	}
}

func activarWorkerSSH(ip string, puerto string){
	fmt.Println("Entramos en activarSSH")
	
	ssh, err := NewSshClient(
		"a800616",
		ip,
		22,
		"/home/a800616/.ssh/id_rsa",
		"")

	fmt.Println("Pasamos crear nuevo cliente")
	
	if err != nil {
		fmt.Printf("SSH init error %v", err)
	} else {
		comando := "/home/a800616/UNI/Tercero/SD/p1-sd-master/master-worker/worker " + ip + " " + puerto
		//comando := "/home/a800616/UNI/Tercero/SD/p1-sd-master/master-worker/worker"
		fmt.Println("Lanzamos comando")
		
		//output, err := ssh.RunCommand(comando)
		ssh.RunCommand(comando)
		
		fmt.Println("Hemos lanzado el comando")
		//fmt.Println(output)
		if err != nil {
			fmt.Printf("SSH run command error %v", err)
		}
	}
}

func main() {
	CONN_TYPE := "tcp"
	
	//HAY QUE HACER UNA LIBRERIA PARA GUARDAR TODAS FUNCIONES DE SSH PORQUE SINO...
	
	//De momento hardcodeamos el vector de rutas a workers:
	workers := []com.Ruta_worker{
		com.Ruta_worker{
			Ip: "155.210.154.195",
			Puerto: "40000",
		},
		com.Ruta_worker{
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
	for i := range workers{
		activarWorkerSSH(workers[i].Ip, workers[i].Puerto)
		fmt.Println("Lanzo el ssh")
		fmt.Println(i)
	}
	//Activamos todos workers con sus correspodientes ips y puertos a escuchar, tambien
	//arrancamos la gorutines que se conectaran con los workers
	for i := range workers{
		go poolGoRutines(chJobs, workers[i].Ip, workers[i].Puerto)
		fmt.Println("Lanzo el gorutine")
		fmt.Println(i)
	}
	fmt.Println("Paso")
	
	fmt.Println(CONN_HOST)
	listener, err := net.Listen(CONN_TYPE, ":" + CONN_PORT)
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
		fmt.Println("He creado el job")
		//Enviamos al canal de las gorutines para que procesen los datos
	    chJobs <- job
		fmt.Println("He enviado el job")
    }
}

