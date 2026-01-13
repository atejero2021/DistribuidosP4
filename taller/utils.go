package main

import (
	"math/rand"
	"sort"
	"time"
)

type ResultadoSimulacion struct {
	Metodo      string
	Test        int
	CochesA     int
	CochesB     int
	CochesC     int
	Duracion    float64
	TotalCoches int
}

type Peticion struct {
	Coche     *Coche
	Respuesta chan bool
}

// Gestor para simular prioridad con canales
type GestorPrioridad struct {
	Entrada   chan Peticion
	Salida    chan bool
	Capacidad int
}

func NewGestorPrioridad(capacidad int) *GestorPrioridad {
	g := &GestorPrioridad{
		Entrada:   make(chan Peticion),
		Salida:    make(chan bool),
		Capacidad: capacidad,
	}
	// Lanzamos la gorutina de control
	go g.loop()
	return g
}

// El coche pide entrar y se bloquea hasta que le den paso
func (g *GestorPrioridad) Entrar(coche *Coche) {
	miCanal := make(chan bool)
	req := Peticion{Coche: coche, Respuesta: miCanal}

	g.Entrada <- req
	<-miCanal
}

// Liberar el recurso
func (g *GestorPrioridad) Salir() {
	g.Salida <- true
}

// Lógica principal de control de cola
func (g *GestorPrioridad) loop() {
	colaEspera := make([]Peticion, 0)
	recursosLibres := g.Capacidad

	for {
		// Si hay sitio y alguien esperando, atendemos
		if recursosLibres > 0 && len(colaEspera) > 0 {
			
			// Reordenamos para colar a los prioritarios
			sort.Slice(colaEspera, func(i, j int) bool {
				if colaEspera[i].Coche.Prioridad != colaEspera[j].Coche.Prioridad {
					return colaEspera[i].Coche.Prioridad < colaEspera[j].Coche.Prioridad
				}
				// Si misma prioridad, orden por ID
				return colaEspera[i].Coche.ID < colaEspera[j].Coche.ID
			})

			// Sacamos al primero
			siguiente := colaEspera[0]
			colaEspera = colaEspera[1:]

			recursosLibres--
			siguiente.Respuesta <- true
			continue
		}

		// Esperar a que alguien llegue o salga
		select {
		case peticion := <-g.Entrada:
			colaEspera = append(colaEspera, peticion)

		case <-g.Salida:
			recursosLibres++
		}
	}
}


func SleepConVariacion(segundos int) {
	msBase := segundos * 1000
	// Variación de 0-500ms
	variacion := rand.Intn(500)

	tiempoTotal := time.Duration(msBase+variacion) * time.Millisecond
	time.Sleep(tiempoTotal)
}

func ordenarPorPrioridad(coches []*Coche) []*Coche {
	// Copia para no tocar el original
	cochesCopia := make([]*Coche, len(coches))
	copy(cochesCopia, coches)

	sort.Slice(cochesCopia, func(i, j int) bool {
		if cochesCopia[i].Prioridad != cochesCopia[j].Prioridad {
			return cochesCopia[i].Prioridad < cochesCopia[j].Prioridad
		}
		return cochesCopia[i].ID < cochesCopia[j].ID
	})

	return cochesCopia
}