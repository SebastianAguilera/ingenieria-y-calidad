package repository_test

import "os"

// getenvOr devuelve la variable de entorno o el valor por defecto si no esta
// definida o esta vacia. Solo para tests de integracion (paquete externo).
func getenvOr(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}
