package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)


func ejecutarTest(t *testing.T, numTest int, config ConfigTaller) {
	
	// Inicializo componentes
	logger := NewLogger()
	gestorPlazas := NewGestorPrioridad(config.NumPlazas)
	gestorMecanicos := NewGestorPrioridad(config.NumMecanicos)

	var wg sync.WaitGroup
	var mu sync.Mutex 
	procesados := 0

	inicio := time.Now()

	procesarCoche := func(coche *Coche) {
		defer wg.Done()

		// Intentar entrar al taller (Plaza)
		gestorPlazas.Entrar(coche)

		// Fase 1: Llegada
		logger.Log(coche, FaseLlegada, Entrando)
		SleepConVariacion(coche.Tiempo)
		logger.Log(coche, FaseLlegada, Saliendo)

		// Fase 2: Reparación (Mecánico)
		gestorMecanicos.Entrar(coche)
		logger.Log(coche, FaseReparacion, Entrando)
		SleepConVariacion(coche.Tiempo)
		logger.Log(coche, FaseReparacion, Saliendo)
		gestorMecanicos.Salir() // Libera mecánico

		// Fase 3: Limpieza
		logger.Log(coche, FaseLimpieza, Entrando)
		SleepConVariacion(coche.Tiempo)
		logger.Log(coche, FaseLimpieza, Saliendo)

		// Fase 4: Revisión
		logger.Log(coche, FaseRevision, Entrando)
		SleepConVariacion(coche.Tiempo)
		logger.Log(coche, FaseRevision, Saliendo)

		// Salir del taller (Libera Plaza)
		gestorPlazas.Salir()

		mu.Lock()
		procesados++
		mu.Unlock()
	}

	idCoche := 1

	// Generador de Coches (A, B, C)
	generar := func(cantidad int, tipo TipoIncidencia, prio Prioridad, tiempo int) {
		for i := 0; i < cantidad; i++ {
			wg.Add(1)
			go procesarCoche(&Coche{
				ID:         idCoche,
				Incidencia: tipo,
				Prioridad:  prio,
				Tiempo:     tiempo,
			})
			idCoche++
			// Pequeña pausa para simular llegada escalonada
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Lanzamos los coches
	generar(config.CochesA, Mecanica, Alta, 5)
	generar(config.CochesB, Electrica, Media, 3)
	generar(config.CochesC, Carroceria, Baja, 1)

	// Esperar finalización
	wg.Wait()
	duracion := time.Since(inicio).Seconds()

	// Validaciones básicas
	totalEsperado := config.CochesA + config.CochesB + config.CochesC
	if procesados != totalEsperado {
		t.Errorf("FALLO: Esperados %d, Procesados %d", totalEsperado, procesados)
	}

	// Resumen
	throughput := float64(procesados) / duracion
	
	fmt.Printf("\nResumen Test %d\n\n", numTest)
	fmt.Printf("Duración: %.2f segundos\n", duracion)
	fmt.Printf("Coches procesados: %d/%d\n", procesados, totalEsperado)
	fmt.Printf("Rendimiento: %.2f coches/seg\n\n", throughput)
}

// TEST 1: Carga Balanceada (10A, 10B, 10C)
func Test1_Config_6Plazas_3Mecanicos(t *testing.T) {
	config := ConfigTaller{NumPlazas: 6, NumMecanicos: 3, CochesA: 10, CochesB: 10, CochesC: 10}
	ejecutarTest(t, 1, config) 
}

func Test2_Config_4Plazas_4Mecanicos(t *testing.T) {
	config := ConfigTaller{NumPlazas: 4, NumMecanicos: 4, CochesA: 10, CochesB: 10, CochesC: 10}
	ejecutarTest(t, 2, config)
}

// TEST 2: Carga Pesada (20A, 5B, 5C)
func Test3_Config_6Plazas_3Mecanicos(t *testing.T) {
	config := ConfigTaller{NumPlazas: 6, NumMecanicos: 3, CochesA: 20, CochesB: 5, CochesC: 5}
	ejecutarTest(t, 3, config)
}

func Test4_Config_4Plazas_4Mecanicos(t *testing.T) {
	config := ConfigTaller{NumPlazas: 4, NumMecanicos: 4, CochesA: 20, CochesB: 5, CochesC: 5}
	ejecutarTest(t, 4, config)
}

// TEST 3: Carga Ligera (5A, 5B, 20C)
func Test5_Config_6Plazas_3Mecanicos(t *testing.T) {
	config := ConfigTaller{NumPlazas: 6, NumMecanicos: 3, CochesA: 5, CochesB: 5, CochesC: 20}
	ejecutarTest(t, 5, config)
}

func Test6_Config_4Plazas_4Mecanicos(t *testing.T) {
	config := ConfigTaller{NumPlazas: 4, NumMecanicos: 4, CochesA: 5, CochesB: 5, CochesC: 20}
	ejecutarTest(t, 6, config)
}

// TestGestorPrioridad verifica que no haya bloqueos (Deadlocks)
func TestGestorPrioridad(t *testing.T) {
	gestor := NewGestorPrioridad(2)
	var wg sync.WaitGroup
	
	coches := []*Coche{
		{ID: 1, Prioridad: Baja},
		{ID: 2, Prioridad: Alta},
		{ID: 3, Prioridad: Media},
	}

	for _, coche := range coches {
		wg.Add(1)
		go func(c *Coche) {
			defer wg.Done()
			gestor.Entrar(c)
			time.Sleep(10 * time.Millisecond)
			gestor.Salir()
		}(coche)
		time.Sleep(5 * time.Millisecond)
	}
	wg.Wait()
}