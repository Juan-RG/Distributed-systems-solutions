/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: client.go
* DESCRIPCIÓN: cliente completo para los cuatro escenarios de la práctica 1
*/
package main

import (
    "fmt"
    "net"
    "com"
    "os"

    "encoding/gob"

)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}


func main(){
    endpoint := "127.0.0.1:30000"
    
    //var num1, num2 int
    //Pedimos el intervalo de numeros en el cual se buscaran los primos
    //fmt.Println("Introduce el numero 1:")
    //fmt.Scanln(&num1)
    //fmt.Println("Introduce el numero 2:")
    //fmt.Scanln(&num2)

    //interval := com.TPInterval{num1, num2}

    interval := com.TPInterval{1000, 70000}

    tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
    checkError(err)

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    checkError(err)

    //modigicar el id para que no sea un hardcorde -> idea:Crear variable global que se vaya incrementando                  ToDo: hacer que el ID sea incrementar o aleatorio
    request := com.Request{0, interval}

    encoder := gob.NewEncoder(conn)
    decoder := gob.NewDecoder(conn)

    err = encoder.Encode(request)
    checkError(err)

    var reply com.Reply
    err = decoder.Decode(&reply)
    checkError(err)

    fmt.Println(reply)
}
