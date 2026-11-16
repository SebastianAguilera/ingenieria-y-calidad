package handler

import "context"

type ServicioPDF interface {
	GenerarPDFProyecto(ctx context.Context, proyectoID uint) (pdf []byte, filename string, err error)
	GenerarPDFSprint(ctx context.Context, proyectoID, sprintID uint) (pdf []byte, filename string, err error)
}
