package main

import (
	"fmt"
	"sync"
	"time"
)

// Constantes de tipos de incidencia
const (
	Mecanica   TipoIncidencia = 0
	Electrica  TipoIncidencia = 1
	Carroceria TipoIncidencia = 2
)

// Constantes de prioridad
const (
	Alta  Prioridad = 0
	Media Prioridad = 1
	Baja  Prioridad = 2
)

// Constantes de fases
const (
	FaseLlegada    Fase = 0
	FaseReparacion Fase = 1
	FaseLimpieza   Fase = 2
	FaseRevision   Fase = 3
)

// Estados
const (
	Entrando EstadoFase = "Entrando"
	Saliendo EstadoFase = "Saliendo"
)

type TipoIncidencia int

func (t TipoIncidencia) String() string {
	switch t {
	case Mecanica:
		return "Mecánica"
	case Electrica:
		return "Eléctrica"
	case Carroceria:
		return "Carrocería"
	default:
		return "Desconocida"
	}
}

type Prioridad int

func (p Prioridad) String() string {
	switch p {
	case Alta:
		return "Alta"
	case Media:
		return "Media"
	case Baja:
		return "Baja"
	default:
		return "Desconocida"
	}
}

type EstadoFase string

type Fase int

func (f Fase) String() string {
	switch f {
	case FaseLlegada:
		return "Llegada y Documentación"
	case FaseReparacion:
		return "Reparación"
	case FaseLimpieza:
		return "Limpieza"
	case FaseRevision:
		return "Revisión Final"
	default:
		return "Desconocida"
	}
}

type Coche struct {
	ID         int
	Incidencia TipoIncidencia
	Prioridad  Prioridad
	Tiempo     int
}

func (c *Coche) String() string {
	return fmt.Sprintf("Coche %d [%s - Prioridad %s]", c.ID, c.Incidencia, c.Prioridad)
}

type Logger struct {
	mu           sync.Mutex
	tiempoInicio time.Time
}

func NewLogger() *Logger {
	return &Logger{
		tiempoInicio: time.Now(),
	}
}

func (l *Logger) Log(coche *Coche, fase Fase, estado EstadoFase) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	tiempoTranscurrido := time.Since(l.tiempoInicio).Seconds()
	fmt.Printf("Tiempo %.2f Coche %d Incidencia %s Fase %s Estado %s\n",
		tiempoTranscurrido, coche.ID, coche.Incidencia, fase, estado)
}

type ConfigTaller struct {
	NumPlazas    int
	NumMecanicos int
	CochesA      int
	CochesB      int
	CochesC      int
}