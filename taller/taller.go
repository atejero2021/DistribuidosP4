/*
ESTE ES EL ÚNICO ARCHIVO QUE SE PUEDE MODIFICAR

RECOMENDACIÓN: Solo modicar a partir de la parte
				donde se encuentran la explicación
				de las otras variables.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	buf    bytes.Buffer
	logger = log.New(&buf, "logger: ", log.Lshortfile)
	msg    string
)

var (
	estadoActual    int
	estadoMutex     sync.RWMutex
	tallerLogger    *Logger
	gestorPlazas    *GestorPrioridad
	gestorMecanicos *GestorPrioridad
	wgCoches        sync.WaitGroup
	inicio          sync.Once
	
	// Canal para bloquear el main hasta terminar
	finPrograma = make(chan bool) 
	
	numPlazas    = 6
	numMecanicos = 3
	cochesA      = 10
	cochesB      = 10
	cochesC      = 10
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()
	
	go func() {
		buf := make([]byte, 512)
		for {
			n, err := conn.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				// Si se cierra la conexión de forma esperada al acabar, salimos
				break
			}
			if n > 0 {
				msg = string(buf[:n])
				/*
					Desde aquí debería salir la información a una goroutine o a una función ordinaria según se requiera
				*/
				
				numero := extraerNumero(msg)
				if numero != -1 {
					actualizarEstado(numero)
					// Arrancamos la generación de coches
					inicio.Do(func() {
						inicializarTaller()
						go generarCoches()
					})
				}
			}
		}
	}()
	
	fmt.Println("Esperando órdenes de la Mutua...")
	
	// Esto bloquea el main indefinidamente hasta que generarCoches diga
	<-finPrograma
	
	fmt.Println("\n=== Taller finalizado ===")
}

func inicializarTaller() {
	tallerLogger = NewLogger()
	gestorPlazas = NewGestorPrioridad(numPlazas)
	gestorMecanicos = NewGestorPrioridad(numMecanicos)
	fmt.Printf("Taller iniciado: %d plazas, %d mecánicos | Coches: A=%d B=%d C=%d\n\n",
		numPlazas, numMecanicos, cochesA, cochesB, cochesC)
}

func extraerNumero(mensaje string) int {
	for _, char := range mensaje {
		if char >= '0' && char <= '9' {
			num, _ := strconv.Atoi(string(char))
			return num
		}
	}
	return -1
}

func actualizarEstado(numero int) {
	estadoMutex.Lock()
	defer estadoMutex.Unlock()
	
	anterior := estadoActual
	if numero >= 0 && numero <= 9 && numero != 7 && numero != 8 {
		estadoActual = numero
		if anterior != estadoActual {
			estados := []string{"Inactivo", "Solo A", "Solo B", "Solo C",
				"Prioridad A", "Prioridad B", "Prioridad C", "", "", "Cerrado"}
			fmt.Printf("\n>>> Estado: %s\n", estados[estadoActual])
		}
	}
}

func generarCoches() {
	idCoche := 1
	pendientesA := cochesA
	pendientesB := cochesB
	pendientesC := cochesC
	
	// Lanzamos los coches
	for i := 0; i < pendientesA; i++ {
		wgCoches.Add(1)
		go procesarCoche(&Coche{ID: idCoche, Incidencia: Mecanica, Prioridad: Alta, Tiempo: 5})
		idCoche++
		time.Sleep(1500 * time.Millisecond)
	}
	
	for i := 0; i < pendientesB; i++ {
		wgCoches.Add(1)
		go procesarCoche(&Coche{ID: idCoche, Incidencia: Electrica, Prioridad: Media, Tiempo: 3})
		idCoche++
		time.Sleep(1500 * time.Millisecond)
	}
	
	for i := 0; i < pendientesC; i++ {
		wgCoches.Add(1)
		go procesarCoche(&Coche{ID: idCoche, Incidencia: Carroceria, Prioridad: Baja, Tiempo: 1})
		idCoche++
		time.Sleep(1500 * time.Millisecond)
	}
	
	// Groutine que espera a que terminen los coches y luego avisa al main para que se cierre.
	go func() {
		wgCoches.Wait()      
		finPrograma <- true 
	}()
}

func procesarCoche(coche *Coche) {
	defer wgCoches.Done()
	
	esperarEstadoValido(coche)
	
	// El coche ocupa la plaza durante todo el proceso
	gestorPlazas.Entrar(coche)
	
	// Fase 1: Llegada y Documentación
	tallerLogger.Log(coche, FaseLlegada, Entrando)
	SleepConVariacion(coche.Tiempo)
	tallerLogger.Log(coche, FaseLlegada, Saliendo)
	
	// Fase 2: Reparación
	gestorMecanicos.Entrar(coche)
	tallerLogger.Log(coche, FaseReparacion, Entrando)
	SleepConVariacion(coche.Tiempo)
	tallerLogger.Log(coche, FaseReparacion, Saliendo)
	gestorMecanicos.Salir()
	
	// Fase 3: Limpieza
	tallerLogger.Log(coche, FaseLimpieza, Entrando)
	SleepConVariacion(coche.Tiempo)
	tallerLogger.Log(coche, FaseLimpieza, Saliendo)
	
	// Fase 4: Revisión Final
	tallerLogger.Log(coche, FaseRevision, Entrando)
	SleepConVariacion(coche.Tiempo)
	tallerLogger.Log(coche, FaseRevision, Saliendo)
	
	// Libera la plaza al salir del taller
	gestorPlazas.Salir()
}

func esperarEstadoValido(coche *Coche) {
	for {
		estadoMutex.RLock()
		estado := estadoActual
		estadoMutex.RUnlock()
		
		if puedeEntrar(coche, estado) {
			return
		}
		
		if estado == 0 || estado == 9 {
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func puedeEntrar(coche *Coche, estado int) bool {
	switch estado {
	case 0, 9:
		return false
	case 1:
		return coche.Incidencia == Mecanica
	case 2:
		return coche.Incidencia == Electrica
	case 3:
		return coche.Incidencia == Carroceria
	case 4, 5, 6:
		return true
	default:
		return true
	}
}