/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*           Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: client.go
* DESCRIPCIÓN: cliente completo para los cuatro escenarios de la práctica 1
*/
package main

import (
    "fmt"
    "time"
    "encoding/gob"
    "com"
    "os"
    "net"
    "sync"
)

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}

// sendRequest envía una petición (id, interval) al servidor. Una petición es un par id 
// (el identificador único de la petición) e interval, el intervalo en el cual se desea que el servidor encuentre los
// números primos. La petición se serializa utilizando el encoder y una vez enviada la petición
// se almacena en una estructura de datos, junto con una estampilla
// temporal. Para evitar condiciones de carrera, la estructura de datos compartida se almacena en una Goroutine
// (handleRequests) y que controla los accesos a través de canales síncronos. En este caso, se añade una nueva
// petición a la estructura de datos mediante el canal addChan
func sendRequest(id int, interval com.TPInterval, endpoint string, wg *sync.WaitGroup){

    //Cambio para realizar x conexiones
    
    
    tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
    checkError(err)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    checkError(err)
    encoder := gob.NewEncoder(conn)
    decoder := gob.NewDecoder(conn)

    addChan := make(chan com.TimeRequest)
    delChan := make(chan com.TimeReply)

    
    go handleRequests(addChan, delChan)
    go receiveReply(decoder, delChan, wg)

    request := com.Request{id, interval}
    timeReq := com.TimeRequest{id, time.Now()}
    err = encoder.Encode(request)

    checkError(err)
    addChan <- timeReq
}

// handleRequests es una Goroutine que garantiza el acceso en exclusión mutua a la tabla de peticiones. La tabla de peticiones
// almacena todas las peticiones activas que se han realizado al servidor y cuándo se han realizado. El objetivo es que el cliente
// pueda calcular, para cada petición, cuál es el tiempo total desde que se envía hasta que se recibe.
// Las peticiones le llegan a la goroutine a través del canal addChan. Por el canal delChan se
// indica que ha llegado una respuesta de una petición. En la respuesta, se obtiene también el timestamp de la recepción.
// Antes de eliminar una petición se imprime por la salida estándar el id de una petición y el tiempo transcurrido, observado
// por el cliente (tiempo de transmisión + tiempo de overheads + tiempo de ejecución efectivo)
func handleRequests(addChan chan com.TimeRequest, delChan chan com.TimeReply) {
    requests := make(map[int]time.Time)
    for {
        select {
            case request := <- addChan:
                requests[request.Id] = request.T
            case reply := <- delChan:
                fmt.Println(reply.Id, " ", reply.T.Sub(requests[reply.Id]))
                delete(requests, reply.Id)
        }
    }
}

// receiveReply recibe las respuestas (id, primos) del servidor. Respuestas que corresponden con peticiones previamente
// realizadas. 
// el encoder y una vez enviada la petición se almacena en una estructura de datos, junto con una estampilla
// temporal. Para evitar condiciones de carrera, la estructura de datos compartida se almacena en una Goroutine
// (handleRequests) y que controla los accesos a través de canales síncronos. En este caso, se añade una nueva
// petición a la estructura de datos mediante el canal addChan
func receiveReply(decoder *gob.Decoder, delChan chan com.TimeReply, wg *sync.WaitGroup){
    var reply com.Reply
    err := decoder.Decode(&reply)
    checkError(err)
    timeReply := com.TimeReply{reply.Id, time.Now()}
    wg.Done()
    delChan <- timeReply
    
}

func main(){
    wg := &sync.WaitGroup{}
    endpoint := "127.0.0.1:40000"
    numIt := 10
    requestTmp := 6
    interval := com.TPInterval{1000, 70000}
    tts := 3000 // time to sleep between consecutive requests

    for i := 0; i < numIt; i++ {
        for t := 1; t <= requestTmp; t++{
            wg.Add(1)
            sendRequest(i * requestTmp + t, interval, endpoint, wg)
           
        }
        time.Sleep(time.Duration(tts) * time.Millisecond)
    }
     //time.Sleep(time.Second * 50)
     wg.Wait()
}