package domain

import "sort"

// HistoriaPriorizada es una historia apta para priorizar junto con los
// indicadores calculados a partir de sus valores de negocio y story points.
type HistoriaPriorizada struct {
	Historia Historia
	ROI      *float64
	WSJF     *float64
}

// CalcularROI devuelve valor_negocio / story_points. Es nil si falta el valor
// de negocio, si no hay story points o si los story points son 0.
func CalcularROI(valorNegocio, storyPoints *float64) *float64 {
	if valorNegocio == nil || storyPoints == nil || *storyPoints == 0 {
		return nil
	}

	roi := *valorNegocio / *storyPoints
	return &roi
}

// CalcularCostoDemora suma los tres valores de priorizacion. Un campo ausente
// cuenta como 0, por lo que siempre devuelve un numero.
func CalcularCostoDemora(valorNegocio, criticidadTiempo, reduccionRiesgo *float64) float64 {
	total := 0.0
	if valorNegocio != nil {
		total += *valorNegocio
	}
	if criticidadTiempo != nil {
		total += *criticidadTiempo
	}
	if reduccionRiesgo != nil {
		total += *reduccionRiesgo
	}
	return total
}

// CalcularWSJF devuelve costoDemora / story_points. Es nil si no hay story
// points o si los story points son 0.
func CalcularWSJF(costoDemora float64, storyPoints *float64) *float64 {
	if storyPoints == nil || *storyPoints == 0 {
		return nil
	}

	wsjf := costoDemora / *storyPoints
	return &wsjf
}

// PriorizarBacklog excluye las historias Hecho y ordena las restantes por WSJF
// descendente. Las historias sin WSJF van al final y los empates se resuelven
// por ID ascendente. No modifica la entrada.
func PriorizarBacklog(historias []Historia) []HistoriaPriorizada {
	priorizadas := make([]HistoriaPriorizada, 0, len(historias))
	for _, historia := range historias {
		if historia.Estado == EstadoHistoriaHecho {
			continue
		}

		costo := CalcularCostoDemora(historia.ValorNegocio, historia.CriticidadTiempo, historia.ReduccionRiesgo)
		priorizadas = append(priorizadas, HistoriaPriorizada{
			Historia: historia,
			ROI:      CalcularROI(historia.ValorNegocio, historia.StoryPoints),
			WSJF:     CalcularWSJF(costo, historia.StoryPoints),
		})
	}

	sort.SliceStable(priorizadas, func(i, j int) bool {
		izquierda, derecha := priorizadas[i], priorizadas[j]

		switch {
		case izquierda.WSJF == nil && derecha.WSJF == nil:
			return izquierda.Historia.ID < derecha.Historia.ID
		case izquierda.WSJF == nil:
			return false
		case derecha.WSJF == nil:
			return true
		case *izquierda.WSJF != *derecha.WSJF:
			return *izquierda.WSJF > *derecha.WSJF
		default:
			return izquierda.Historia.ID < derecha.Historia.ID
		}
	})

	return priorizadas
}
