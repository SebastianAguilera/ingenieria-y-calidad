package service

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"ingenieria-y-calidad/internal/domain"
)

// PDFService genera PDFs del reporte reutilizando ReporteService (US-22).
// No recalcula ninguna métrica.
type PDFService struct {
	reportes *ReporteService
}

func NewPDFService(reportes *ReporteService) *PDFService {
	return &PDFService{reportes: reportes}
}

func (s *PDFService) GenerarPDFProyecto(ctx context.Context, proyectoID uint) ([]byte, string, error) {
	reporte, err := s.reportes.ObtenerReporteProyecto(ctx, proyectoID)
	if err != nil {
		return nil, "", err
	}
	fecha := time.Now().Format("2006-01-02")
	filename := nombreArchivoProyecto(reporte.Nombre, fecha)
	buf, err := generarPDFProyecto(reporte)
	if err != nil {
		return nil, "", fmt.Errorf("generando PDF de proyecto: %w", err)
	}
	return buf, filename, nil
}

func (s *PDFService) GenerarPDFSprint(ctx context.Context, proyectoID, sprintID uint) ([]byte, string, error) {
	reporte, err := s.reportes.ObtenerReporteSprint(ctx, proyectoID, sprintID)
	if err != nil {
		return nil, "", err
	}
	fecha := time.Now().Format("2006-01-02")
	filename := nombreArchivoSprint(fmt.Sprintf("proyecto-%d", reporte.ProyectoID), reporte.Identificador, fecha)
	buf, err := generarPDFSprint(reporte)
	if err != nil {
		return nil, "", fmt.Errorf("generando PDF de sprint: %w", err)
	}
	return buf, filename, nil
}

// ---------------------------------------------------------------------------
// Helpers de nombre de archivo
// ---------------------------------------------------------------------------

func nombreArchivoProyecto(nombreProyecto, fecha string) string {
	return fmt.Sprintf("reporte-%s-proyecto-%s.pdf", slugify(nombreProyecto), fecha)
}

func nombreArchivoSprint(slugProyecto, identificadorSprint, fecha string) string {
	return fmt.Sprintf("reporte-%s-sprint-%s-%s.pdf", slugify(slugProyecto), slugify(identificadorSprint), fecha)
}

// slugify convierte una cadena en un slug ASCII seguro para nombres de archivo.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen && b.Len() > 0 {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	result := strings.TrimRight(b.String(), "-")
	if result == "" {
		return "reporte"
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// ---------------------------------------------------------------------------
// Generadores de PDF puros (sin dependencias de infraestructura)
// ---------------------------------------------------------------------------

func generarPDFProyecto(r domain.ReporteProyecto) ([]byte, error) {
	f := nuevoPDF()
	tr := f.UnicodeTranslatorFromDescriptor("")
	f.AddPage()

	addTitulo(f, tr, "Reporte de Proyecto")
	addSeparador(f)

	addSeccion(f, tr, "Encabezado")
	addFila(f, tr, "Nombre", r.Nombre)
	addFila(f, tr, "Descripcion", r.Descripcion)
	addFila(f, tr, "Fecha Inicio", r.FechaInicio.Format("2006-01-02"))
	if r.FechaFin != nil {
		addFila(f, tr, "Fecha Fin", r.FechaFin.Format("2006-01-02"))
	}
	addFila(f, tr, "Estado", string(r.Estado))
	addFila(f, tr, "Base", r.Base)
	addSeparador(f)

	addSeccion(f, tr, "Metricas")
	addMetricaFloat(f, tr, "Velocidad", r.Metricas.Velocidad)
	addMetricaFloat(f, tr, "Desviacion de Esfuerzo", r.Metricas.DesviacionEsfuerzo)
	addMetricaFloat(f, tr, "Porcentaje Historias Completadas", r.Metricas.PorcentajeHistoriasCompletadas)
	addFila(f, tr, "Horas Estimadas", strconv.FormatFloat(r.Metricas.HorasEstimadas, 'f', 2, 64))
	addFila(f, tr, "Horas Reales", strconv.FormatFloat(r.Metricas.HorasReales, 'f', 2, 64))
	addFila(f, tr, "Defectos Detectados", strconv.Itoa(r.Metricas.DefectosDetectados))
	addFila(f, tr, "Defectos Resueltos", strconv.Itoa(r.Metricas.DefectosResueltos))
	addMetricaFloat(f, tr, "Ratio de Defectos", r.Metricas.RatioDefectos)
	addMetricaFloat(f, tr, "Sprints Restantes", r.Metricas.SprintsRestantes)
	addFila(f, tr, "WIP", strconv.Itoa(r.Metricas.WIP))
	addMetricaFloat(f, tr, "Tasa de Entrega", r.Metricas.TasaEntrega)
	addMetricaFloat(f, tr, "Lead Time Estimado", r.Metricas.LeadTimeEstimado)
	addFila(f, tr, "Defectos Abiertos", strconv.Itoa(r.Metricas.DefectosAbiertos))
	addSeparador(f)

	addSeccion(f, tr, "Historias")
	if len(r.Historias) == 0 {
		addFila(f, tr, "", "Sin datos")
	} else {
		addCabeceraTablaHistorias(f, tr)
		for _, h := range r.Historias {
			addFilaHistoria(f, tr, h)
		}
	}
	addSeparador(f)

	addSeccion(f, tr, "Defectos")
	if len(r.Defectos) == 0 {
		addFila(f, tr, "", "Sin datos")
	} else {
		addCabeceraTablaDefectos(f, tr)
		for _, d := range r.Defectos {
			addFilaDefecto(f, tr, d)
		}
	}

	return outputPDF(f)
}

func generarPDFSprint(r domain.ReporteSprint) ([]byte, error) {
	f := nuevoPDF()
	tr := f.UnicodeTranslatorFromDescriptor("")
	f.AddPage()

	addTitulo(f, tr, "Reporte de Sprint")
	addSeparador(f)

	addSeccion(f, tr, "Encabezado")
	addFila(f, tr, "Identificador", r.Identificador)
	addFila(f, tr, "Sprint Goal", r.SprintGoal)
	addFila(f, tr, "Fecha Inicio", r.FechaInicio.Format("2006-01-02"))
	addFila(f, tr, "Fecha Fin", r.FechaFin.Format("2006-01-02"))
	addFila(f, tr, "Estado", string(r.Estado))
	addFila(f, tr, "Base", r.Base)
	addSeparador(f)

	addSeccion(f, tr, "Metricas")
	addFila(f, tr, "SP Planificados", strconv.FormatFloat(r.Metricas.StoryPointsPlanificados, 'f', 2, 64))
	addFila(f, tr, "SP Completados", strconv.FormatFloat(r.Metricas.StoryPointsCompletados, 'f', 2, 64))
	addMetricaFloat(f, tr, "Porcentaje Completado", r.Metricas.PorcentajeCompletado)
	addFila(f, tr, "Historias Completadas", strconv.Itoa(r.Metricas.HistoriasCompletadas))
	addFila(f, tr, "Horas Estimadas", strconv.FormatFloat(r.Metricas.HorasEstimadas, 'f', 2, 64))
	addFila(f, tr, "Horas Reales", strconv.FormatFloat(r.Metricas.HorasReales, 'f', 2, 64))
	addMetricaFloat(f, tr, "Desviacion de Esfuerzo", r.Metricas.DesviacionEsfuerzo)
	addFila(f, tr, "Defectos Detectados", strconv.Itoa(r.Metricas.DefectosDetectados))
	addFila(f, tr, "Defectos Resueltos", strconv.Itoa(r.Metricas.DefectosResueltos))
	addSeparador(f)

	addSeccion(f, tr, "Historias")
	if len(r.Historias) == 0 {
		addFila(f, tr, "", "Sin datos")
	} else {
		addCabeceraTablaHistorias(f, tr)
		for _, h := range r.Historias {
			addFilaHistoria(f, tr, h)
		}
	}
	addSeparador(f)

	addSeccion(f, tr, "Defectos")
	if len(r.Defectos) == 0 {
		addFila(f, tr, "", "Sin datos")
	} else {
		addCabeceraTablaDefectos(f, tr)
		for _, d := range r.Defectos {
			addFilaDefecto(f, tr, d)
		}
	}

	return outputPDF(f)
}

// ---------------------------------------------------------------------------
// Primitivas de construcción de PDF
// ---------------------------------------------------------------------------

func nuevoPDF() *fpdf.Fpdf {
	f := fpdf.New("P", "mm", "A4", "")
	f.SetMargins(15, 15, 15)
	f.SetAutoPageBreak(true, 15)
	f.SetCompression(false)
	return f
}

func addTitulo(f *fpdf.Fpdf, tr func(string) string, titulo string) {
	f.SetFont("Helvetica", "B", 16)
	f.SetTextColor(30, 30, 100)
	f.Cell(0, 10, tr(titulo))
	f.Ln(12)
	f.SetTextColor(0, 0, 0)
}

func addSeccion(f *fpdf.Fpdf, tr func(string) string, nombre string) {
	f.SetFont("Helvetica", "B", 12)
	f.SetFillColor(220, 230, 250)
	f.SetTextColor(20, 20, 80)
	f.CellFormat(0, 7, tr(nombre), "", 1, "L", true, 0, "")
	f.SetTextColor(0, 0, 0)
	f.SetFillColor(255, 255, 255)
	f.Ln(2)
}

func addSeparador(f *fpdf.Fpdf) {
	f.Ln(4)
}

func addFila(f *fpdf.Fpdf, tr func(string) string, clave, valor string) {
	f.SetFont("Helvetica", "B", 9)
	f.CellFormat(60, 6, tr(clave+":"), "", 0, "L", false, 0, "")
	f.SetFont("Helvetica", "", 9)
	f.MultiCell(0, 6, tr(valor), "", "L", false)
}

func addMetricaFloat(f *fpdf.Fpdf, tr func(string) string, nombre string, v *float64) {
	var valor string
	if v == nil {
		valor = "Sin datos"
	} else {
		valor = strconv.FormatFloat(*v, 'f', 2, 64)
	}
	addFila(f, tr, nombre, valor)
}

func addCabeceraTablaHistorias(f *fpdf.Fpdf, tr func(string) string) {
	f.SetFont("Helvetica", "B", 9)
	f.SetFillColor(200, 210, 240)
	anchos := []float64{20, 60, 25, 15, 25, 25}
	cabeceras := []string{"Codigo", "Titulo", "Estado", "SP", "Hrs Est.", "Hrs Reales"}
	for i, cab := range cabeceras {
		f.CellFormat(anchos[i], 6, tr(cab), "1", 0, "C", true, 0, "")
	}
	f.Ln(-1)
	f.SetFillColor(255, 255, 255)
}

func addFilaHistoria(f *fpdf.Fpdf, tr func(string) string, h domain.Historia) {
	f.SetFont("Helvetica", "", 8)
	anchos := []float64{20, 60, 25, 15, 25, 25}
	sp := "Sin datos"
	if h.StoryPoints != nil {
		sp = strconv.FormatFloat(*h.StoryPoints, 'f', 0, 64)
	}
	hEst := "Sin datos"
	if h.HorasEstimadas != nil {
		hEst = strconv.FormatFloat(*h.HorasEstimadas, 'f', 2, 64)
	}
	hReal := "Sin datos"
	if h.EsfuerzoReal != nil {
		hReal = strconv.FormatFloat(*h.EsfuerzoReal, 'f', 2, 64)
	}
	valores := []string{h.Codigo, h.Titulo, string(h.Estado), sp, hEst, hReal}
	for i, v := range valores {
		f.CellFormat(anchos[i], 6, tr(v), "1", 0, "L", false, 0, "")
	}
	f.Ln(-1)
}

func addCabeceraTablaDefectos(f *fpdf.Fpdf, tr func(string) string) {
	f.SetFont("Helvetica", "B", 9)
	f.SetFillColor(200, 210, 240)
	anchos := []float64{20, 70, 25, 25, 25, 25}
	cabeceras := []string{"Codigo", "Descripcion", "Severidad", "Estado", "Sprint Det.", "Sprint Res."}
	for i, cab := range cabeceras {
		f.CellFormat(anchos[i], 6, tr(cab), "1", 0, "C", true, 0, "")
	}
	f.Ln(-1)
	f.SetFillColor(255, 255, 255)
}

func addFilaDefecto(f *fpdf.Fpdf, tr func(string) string, d domain.Defecto) {
	f.SetFont("Helvetica", "", 8)
	anchos := []float64{20, 70, 25, 25, 25, 25}
	codigo := fmt.Sprintf("DEF-%d", d.ID)
	sprintRes := "Sin datos"
	if d.SprintResolucionID != nil {
		sprintRes = strconv.FormatUint(uint64(*d.SprintResolucionID), 10)
	}
	valores := []string{
		codigo,
		d.Descripcion,
		string(d.Severidad),
		string(d.Estado),
		strconv.FormatUint(uint64(d.SprintDeteccionID), 10),
		sprintRes,
	}
	for i, v := range valores {
		f.CellFormat(anchos[i], 6, tr(v), "1", 0, "L", false, 0, "")
	}
	f.Ln(-1)
}

func outputPDF(f *fpdf.Fpdf) ([]byte, error) {
	if err := f.Error(); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := f.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
