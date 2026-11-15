package domain

import (
	"math"
	"sort"
	"strconv"
)

// CalcularDispersionPoker calcula las estadísticas de una ronda. La desviación
// estándar es poblacional: se divide por la cantidad de votos, no por n-1.
func CalcularDispersionPoker(votos []VotoPoker) DispersionPoker {
	dispersion := DispersionPoker{
		TotalVotos:    len(votos),
		VotosPorValor: map[string]int{},
	}
	if len(votos) == 0 {
		return dispersion
	}

	valores := make([]float64, len(votos))
	var suma float64
	for idx, voto := range votos {
		valores[idx] = voto.Puntos
		suma += voto.Puntos
		dispersion.VotosPorValor[formatearPunto(voto.Puntos)]++
	}

	dispersion.Promedio = suma / float64(len(votos))

	sort.Float64s(valores)
	dispersion.Minimo = valores[0]
	dispersion.Maximo = valores[len(valores)-1]
	dispersion.Rango = dispersion.Maximo - dispersion.Minimo
	dispersion.HayDispersion = dispersion.Maximo != dispersion.Minimo

	medio := len(valores) / 2
	if len(valores)%2 == 1 {
		dispersion.Mediana = valores[medio]
	} else {
		dispersion.Mediana = (valores[medio-1] + valores[medio]) / 2
	}

	var sumaCuadrados float64
	for _, valor := range valores {
		diferencia := valor - dispersion.Promedio
		sumaCuadrados += diferencia * diferencia
	}
	dispersion.DesviacionEstandar = math.Sqrt(sumaCuadrados / float64(len(valores)))

	return dispersion
}

func formatearPunto(puntos float64) string {
	return strconv.FormatFloat(puntos, 'f', -1, 64)
}
