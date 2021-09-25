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
    "os"
    "net"
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

//Funcion para pasar a [] bytes
func EncodeToBytes(p interface{}) []byte {

    buf := bytes.Buffer{}
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(p)
    if err != nil {
        log.Fatal(err)
    }
    //devuelvo un struct como bytes
    return buf.Bytes()
}

//funcion para comprimir
func Compress(s []byte) []byte {

    zipbuf := bytes.Buffer{}
    zipped := gzip.NewWriter(&zipbuf)
    zipped.Write(s)
    zipped.Close()
    //comprimo el array de bytes para enviar menos informacion
    return zipbuf.Bytes()
}

//post return a RePLY
func Decode(data []byte) com.Reply {

    r := com.Reply{}
    dec := gob.NewDecoder(bytes.NewReader(data))
    err := dec.Decode(&r)
    if err != nil {
        log.Fatal(err)
    }
    //Devuelvo un reply decodificado de un array de bytes
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
    //descomprimo un array de bytes comprimido
    return data
}

func main(){
    endpoint := "127.0.0.1:30000"
    
    var num1, num2 int
    //Pedimos el intervalo de numeros en el cual se buscaran los primos
    fmt.Println("Introduce el numero 1:")
    fmt.Scanln(&num1)
    fmt.Println("Introduce el numero 2:")
    fmt.Scanln(&num2)

    interval := com.TPInterval{num1, num2}

    tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
    checkError(err)

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    checkError(err)

    //modigicar el id para que no sea un hardcorde -> idea:Crear variable global que se vaya incrementando                  ToDo: hacer que el ID sea incrementar o aleatorio
    request := com.Request{0, interval}


    //paso el struct a bytes
    dataOut := EncodeToBytes(request)
    //Lo comprimo para enviar menos bytes
    dataOut = Compress(dataOut)
    //Lo escribo en el socket
    conn.Write(dataOut)

    //espero la respuesta
    //Buff of 1024  because TCP only contain up to 65495 bytes of payload
    buf := make([]byte, 1024)
    //mensaje completo
    var allMsg []byte
    //bytes recibidos
    total := 0
    for {
        //leo del socket
        n, err := conn.Read(buf)
        if err != nil {
            //Cambiar por una condicion de for el break no me gusta
            break
        }
        //Acumulo los bytes totales recibidos
        total += n
        //forma de crear un array dinamico en goland
        allMsg = append(allMsg, buf...)
    }

    //Descomprimo los datos recibidos de 0 -> N(numero de bytes recibidos)
    dataIn := Decompress(allMsg[0:total])
    //println(dataIn)
    //lo decodifico para obtener la respuesta
    reply := Decode(dataIn)
    
    fmt.Println(reply)

}
