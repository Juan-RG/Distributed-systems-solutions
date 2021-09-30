/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: definitions.go
* DESCRIPCIÓN: contiene las definiciones de estructuras de datos necesarias para
*			la práctica 1
*/
package com

import "time"
import "net"

type TPInterval struct {
    A int
    B int
}

type Request struct {
    Id int
    Interval TPInterval
}

type Job struct {
    Conn net.Conn
    Request Request
}

type TimeRequest struct {
    Id int
    T time.Time
}

type Reply struct {
    Id int
    Primes []int
}

type TimeReply struct {
    Id int
    T time.Time
}

type ruta_worker struct {
    ip string
    puerto string
}