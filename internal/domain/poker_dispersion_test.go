package domain

import (
	"math"
	"testing"
	"time"
)

func votosConPuntos(puntos ...float64) []VotoPoker {
	votos := make([]VotoPoker, len(puntos))
	for idx, valor := range puntos {
		votos[idx] = VotoPoker{Puntos: valor}
	}
	return votos
}

func TestCalcularDispersionPoker(t *testing.T) {
	casos := []struct {
		nombre           string
		puntos           []float64
		promedio         float64
		mediana          float64
		minimo           float64
		maximo           float64
		rango            float64
		desviacion       float64
		hayDispersion    bool
		totalVotos       int
		votosPorValor    map[string]int
	}{
		{
			nombre:        "votos 1, 3 y 8",
			puntos:        []float64{1, 3, 8},
			promedio:      4,
			mediana:       3,
			minimo:        1,
			maximo:        8,
			rango:         7,
			desviacion:    math.Sqrt(26.0 / 3.0),
			hayDispersion: true,
			totalVotos:    3,
			votosPorValor: map[string]int{"1": 1, "3": 1, "8": 1},
		},
		{
			nombre:        "votos iguales 3, 3 y 3",
			puntos:        []float64{3, 3, 3},
			promedio:      3,
			mediana:       3,
			minimo:        3,
			maximo:        3,
			rango:         0,
			desviacion:    0,
			hayDispersion: false,
			totalVotos:    3,
			votosPorValor: map[string]int{"3": 3},
		},
		{
			nombre:        "votos distintos 1, 2 y 5",
			puntos:        []float64{1, 2, 5},
			promedio:      8.0 / 3.0,
			mediana:       2,
			minimo:        1,
			maximo:        5,
			rango:         4,
			desviacion:    math.Sqrt(26.0 / 9.0),
			hayDispersion: true,
			totalVotos:    3,
			votosPorValor: map[string]int{"1": 1, "2": 1, "5": 1},
		},
		{
			nombre:        "un unico voto",
			puntos:        []float64{5},
			promedio:      5,
			mediana:       5,
			minimo:        5,
			maximo:        5,
			rango:         0,
			desviacion:    0,
			hayDispersion: false,
			totalVotos:    1,
			votosPorValor: map[string]int{"5": 1},
		},
		{
			nombre:        "cantidad par de votos usa el promedio central",
			puntos:        []float64{1, 2, 3, 4},
			promedio:      2.5,
			mediana:       2.5,
			minimo:        1,
			maximo:        4,
			rango:         3,
			desviacion:    math.Sqrt(1.25),
			hayDispersion: true,
			totalVotos:    4,
			votosPorValor: map[string]int{"1": 1, "2": 1, "3": 1, "4": 1},
		},
		{
			nombre:        "votos decimales sin dispersion",
			puntos:        []float64{0.5, 0.5},
			promedio:      0.5,
			mediana:       0.5,
			minimo:        0.5,
			maximo:        0.5,
			rango:         0,
			desviacion:    0,
			hayDispersion: false,
			totalVotos:    2,
			votosPorValor: map[string]int{"0.5": 2},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			dispersion := CalcularDispersionPoker(votosConPuntos(caso.puntos...))

			if math.Abs(dispersion.Promedio-caso.promedio) > 1e-9 {
				t.Fatalf("Promedio = %v, se esperaba %v", dispersion.Promedio, caso.promedio)
			}
			if math.Abs(dispersion.Mediana-caso.mediana) > 1e-9 {
				t.Fatalf("Mediana = %v, se esperaba %v", dispersion.Mediana, caso.mediana)
			}
			if dispersion.Minimo != caso.minimo {
				t.Fatalf("Minimo = %v, se esperaba %v", dispersion.Minimo, caso.minimo)
			}
			if dispersion.Maximo != caso.maximo {
				t.Fatalf("Maximo = %v, se esperaba %v", dispersion.Maximo, caso.maximo)
			}
			if dispersion.Rango != caso.rango {
				t.Fatalf("Rango = %v, se esperaba %v", dispersion.Rango, caso.rango)
			}
			if math.Abs(dispersion.DesviacionEstandar-caso.desviacion) > 1e-9 {
				t.Fatalf("DesviacionEstandar = %v, se esperaba %v", dispersion.DesviacionEstandar, caso.desviacion)
			}
			if dispersion.HayDispersion != caso.hayDispersion {
				t.Fatalf("HayDispersion = %v, se esperaba %v", dispersion.HayDispersion, caso.hayDispersion)
			}
			if dispersion.TotalVotos != caso.totalVotos {
				t.Fatalf("TotalVotos = %v, se esperaba %v", dispersion.TotalVotos, caso.totalVotos)
			}
			if len(dispersion.VotosPorValor) != len(caso.votosPorValor) {
				t.Fatalf("VotosPorValor = %v, se esperaba %v", dispersion.VotosPorValor, caso.votosPorValor)
			}
			for valor, cantidad := range caso.votosPorValor {
				if dispersion.VotosPorValor[valor] != cantidad {
					t.Fatalf("VotosPorValor[%q] = %d, se esperaba %d",
						valor, dispersion.VotosPorValor[valor], cantidad)
				}
			}
		})
	}
}

func TestCalcularDispersionPokerSinVotos(t *testing.T) {
	dispersion := CalcularDispersionPoker(nil)

	if dispersion.TotalVotos != 0 {
		t.Fatalf("TotalVotos = %d, se esperaba 0", dispersion.TotalVotos)
	}
	if dispersion.HayDispersion {
		t.Fatal("HayDispersion debe ser false sin votos")
	}
	if len(dispersion.VotosPorValor) != 0 {
		t.Fatalf("VotosPorValor = %v, se esperaba vacio", dispersion.VotosPorValor)
	}
}

func TestEstadoRondaPoker(t *testing.T) {
	votando := RondaPoker{NumeroRonda: 1}
	if obtenido := votando.EstadoRonda(); obtenido != EstadoSesionVotando {
		t.Fatalf("EstadoRonda() = %v, se esperaba %v", obtenido, EstadoSesionVotando)
	}

	reveladaEn := time.Now()
	revelada := RondaPoker{NumeroRonda: 1, ReveladaEn: &reveladaEn}
	if obtenido := revelada.EstadoRonda(); obtenido != EstadoSesionRevelada {
		t.Fatalf("EstadoRonda() = %v, se esperaba %v", obtenido, EstadoSesionRevelada)
	}
}
