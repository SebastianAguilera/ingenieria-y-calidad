package domain

type MetricasSprint struct {
	StoryPointsPlanificados float64
	StoryPointsCompletados  float64
	PorcentajeCompletado    *float64
	HistoriasCompletadas    int
	HorasEstimadas          float64
	HorasReales             float64
	DesviacionEsfuerzo      *float64
	DefectosDetectados      int
	DefectosResueltos       int
}

type MetricasProyecto struct {
	Velocidad                      *float64
	PorcentajeHistoriasCompletadas *float64
	HorasEstimadas                 float64
	HorasReales                    float64
	DesviacionEsfuerzo             *float64
	DefectosDetectados             int
	DefectosResueltos              int
	RatioDefectos                  *float64
}

func CalcularMetricasSprint(historias []Historia, defectos []Defecto, sprintID uint) MetricasSprint {
	planificados := CalcularSPPlanificados(historias)
	completados := CalcularSPCompletados(historias)

	return MetricasSprint{
		StoryPointsPlanificados: planificados,
		StoryPointsCompletados:  completados,
		PorcentajeCompletado:    CalcularPorcentajeCompletado(completados, planificados),
		HistoriasCompletadas:    ContarHistoriasCompletadas(historias),
		HorasEstimadas:          SumarHorasEstimadas(historias),
		HorasReales:             SumarHorasReales(historias),
		DesviacionEsfuerzo:      CalcularDesviacionEsfuerzo(historias),
		DefectosDetectados:      contarDefectosDetectados(defectos, sprintID),
		DefectosResueltos:       contarDefectosResueltosEnSprint(defectos, sprintID),
	}
}

func CalcularMetricasProyecto(sprints []Sprint, asignaciones []SprintBacklogItem, historias []Historia, defectos []Defecto) MetricasProyecto {
	historiasCompletadas := ContarHistoriasCompletadas(historias)

	return MetricasProyecto{
		Velocidad:                      CalcularVelocidad(sprints, asignaciones, historias),
		PorcentajeHistoriasCompletadas: CalcularPorcentajeCompletado(float64(historiasCompletadas), float64(len(historias))),
		HorasEstimadas:                 SumarHorasEstimadas(historias),
		HorasReales:                    SumarHorasReales(historias),
		DesviacionEsfuerzo:             CalcularDesviacionEsfuerzo(historias),
		DefectosDetectados:             len(defectos),
		DefectosResueltos:              contarDefectosResueltos(defectos),
		RatioDefectos:                  CalcularRatioDefectos(defectos, historiasCompletadas),
	}
}

func CalcularSPPlanificados(historias []Historia) float64 {
	total := 0.0
	for _, historia := range historias {
		total += valorStoryPoints(historia.StoryPoints)
	}
	return total
}

func CalcularSPCompletados(historias []Historia) float64 {
	total := 0.0
	for _, historia := range historias {
		if historia.Estado == EstadoHistoriaHecho {
			total += valorStoryPoints(historia.StoryPoints)
		}
	}
	return total
}

func CalcularPorcentajeCompletado(completados, planificados float64) *float64 {
	if planificados == 0 {
		return nil
	}

	porcentaje := completados / planificados
	return &porcentaje
}

func ContarHistoriasCompletadas(historias []Historia) int {
	total := 0
	for _, historia := range historias {
		if historia.Estado == EstadoHistoriaHecho {
			total++
		}
	}
	return total
}

func SumarHorasEstimadas(historias []Historia) float64 {
	total := 0.0
	for _, historia := range historias {
		if historia.HorasEstimadas != nil {
			total += *historia.HorasEstimadas
		}
	}
	return total
}

func SumarHorasReales(historias []Historia) float64 {
	total := 0.0
	for _, historia := range historias {
		if historia.EsfuerzoReal != nil {
			total += *historia.EsfuerzoReal
		}
	}
	return total
}

func CalcularDesviacionEsfuerzo(historias []Historia) *float64 {
	var horasEstimadas, horasReales float64
	pares := 0

	for _, historia := range historias {
		if historia.HorasEstimadas == nil || historia.EsfuerzoReal == nil {
			continue
		}
		horasEstimadas += *historia.HorasEstimadas
		horasReales += *historia.EsfuerzoReal
		pares++
	}

	if pares == 0 || horasEstimadas == 0 {
		return nil
	}

	desviacion := (horasReales - horasEstimadas) / horasEstimadas
	return &desviacion
}

func CalcularVelocidad(sprints []Sprint, asignaciones []SprintBacklogItem, historias []Historia) *float64 {
	historiasPorID := make(map[uint]Historia, len(historias))
	for _, historia := range historias {
		historiasPorID[historia.ID] = historia
	}

	asignacionesPorSprint := make(map[uint][]uint)
	for _, asignacion := range asignaciones {
		asignacionesPorSprint[asignacion.SprintID] = append(
			asignacionesPorSprint[asignacion.SprintID],
			asignacion.HistoriaID,
		)
	}

	total := 0.0
	cerrados := 0

	for _, sprint := range sprints {
		if sprint.Estado != EstadoSprintCerrado {
			continue
		}
		cerrados++

		for _, historiaID := range asignacionesPorSprint[sprint.ID] {
			historia, existe := historiasPorID[historiaID]
			if !existe || historia.Estado != EstadoHistoriaHecho {
				continue
			}
			total += valorStoryPoints(historia.StoryPoints)
		}
	}

	if cerrados == 0 {
		return nil
	}

	velocidad := total / float64(cerrados)
	return &velocidad
}

func CalcularRatioDefectos(defectos []Defecto, historiasCompletadas int) *float64 {
	if historiasCompletadas == 0 {
		return nil
	}

	detectados := 0
	for _, defecto := range defectos {
		if defecto.Estado != EstadoDefectoRechazado {
			detectados++
		}
	}

	ratio := float64(detectados) / float64(historiasCompletadas)
	return &ratio
}

func contarDefectosDetectados(defectos []Defecto, sprintID uint) int {
	total := 0
	for _, defecto := range defectos {
		if defecto.SprintDeteccionID == sprintID {
			total++
		}
	}
	return total
}

func contarDefectosResueltosEnSprint(defectos []Defecto, sprintID uint) int {
	total := 0
	for _, defecto := range defectos {
		if defecto.SprintResolucionID != nil &&
			*defecto.SprintResolucionID == sprintID &&
			defecto.Estado.EsDeResolucion() {
			total++
		}
	}
	return total
}

func contarDefectosResueltos(defectos []Defecto) int {
	total := 0
	for _, defecto := range defectos {
		if defecto.Estado.EsDeResolucion() {
			total++
		}
	}
	return total
}

func valorStoryPoints(storyPoints *float64) float64 {
	if storyPoints == nil {
		return 0
	}
	return *storyPoints
}
