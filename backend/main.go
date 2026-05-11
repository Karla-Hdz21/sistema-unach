package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jung-kurt/gofpdf"
	_ "github.com/lib/pq"
	"gopkg.in/gomail.v2"
)

type AsistenciaRequest struct {
	AlumnoID string `json:"alumno_id"`
	ClaseID  int    `json:"clase_id"`
}

func avisarAPython(alumnoID string, materia string) {
	url := "http://3.84.127.65:8001/notificar" 
	datos := map[string]string{"alumno_id": alumnoID, "materia": materia}
	body, _ := json.Marshal(datos)
	http.Post(url, "application/json", bytes.NewBuffer(body))
}

func esHorarioPermitido(claseNombre string) bool {
	ahora := time.Now()
	dia := ahora.Weekday()
	minutos := ahora.Hour()*60 + ahora.Minute()

	if claseNombre == "Compiladores" {
		if dia == time.Monday && minutos >= 1080 && minutos <= 1140 { return true }
		if dia == time.Wednesday && minutos >= 960 && minutos <= 1080 { return true }
		if dia == time.Friday && minutos >= 1020 && minutos <= 1140 { return true }
	}
	if claseNombre == "Taller 4" {
		if dia == time.Tuesday && minutos >= 1140 && minutos <= 1200 { return true }
		if dia == time.Thursday && minutos >= 1200 && minutos <= 1260 { return true }
	}
	if dia == time.Saturday || dia == time.Sunday { return true }
	return false
}

func main() {
	// Asegúrate de que la base de datos esté corriendo en el puerto 5432
	connStr := "postgresql://postgres:unah2026@db:5432/sistema_unach?sslmode=disable"
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://sistema-unach.vercel.app, http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// RUTA PARA MATERIAS: El DISTINCT evita que se repitan en la interfaz
	handlerMaterias := func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT DISTINCT id, nombre FROM clases ORDER BY nombre ASC")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error de base de datos"})
		}
		defer rows.Close()

		var clases []fiber.Map
		for rows.Next() {
			var id int
			var nom string
			rows.Scan(&id, &nom)
			clases = append(clases, fiber.Map{"id": id, "nombre": nom})
		}
		return c.JSON(clases)
	}

	app.Get("/clases", handlerMaterias)
	app.Get("/api/materias", handlerMaterias)

	// RUTA PARA ASISTENCIA
	app.Post("/api/asistencia", func(c *fiber.Ctx) error {
		var req AsistenciaRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Datos inválidos"})
		}
		
		var nombreClase string
		db.QueryRow("SELECT nombre FROM clases WHERE id = $1", req.ClaseID).Scan(&nombreClase)
		
		if !esHorarioPermitido(nombreClase) {
			return c.Status(403).JSON(fiber.Map{"mensaje": "Fuera de horario"})
		}

		db.Exec("INSERT INTO asistencias (alumno_id, clase_id, fecha_hora) VALUES ($1, $2, NOW())", req.AlumnoID, req.ClaseID)
		
		var nombreAlum string
		db.QueryRow("SELECT nombre FROM alumnos WHERE id = $1", req.AlumnoID).Scan(&nombreAlum)

		go avisarAPython(req.AlumnoID, nombreClase)
		return c.JSON(fiber.Map{"mensaje": "¡Bienvenido, " + nombreAlum + "!"})
	})

	// RUTA PARA GENERAR PDF Y ENVIAR CORREO
	app.Post("/api/cerrar-clase", func(c *fiber.Ctx) error {
		var req struct { ClaseID int `json:"clase_id"` }
		c.BodyParser(&req)

		var nombreClase string
		db.QueryRow("SELECT nombre FROM clases WHERE id = $1", req.ClaseID).Scan(&nombreClase)

		rows, _ := db.Query(`SELECT a.nombre, CASE WHEN asis.id IS NULL THEN 'Faltó' ELSE 'Asistió' END 
                             FROM alumnos a LEFT JOIN asistencias asis ON a.id = asis.alumno_id 
                             AND asis.clase_id = $1 AND asis.fecha_hora::date = CURRENT_DATE`, req.ClaseID)
		
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, "REPORTE DE ASISTENCIA - " + nombreClase)
		pdf.Ln(12)

		for rows.Next() {
			var n, e string
			rows.Scan(&n, &e)
			pdf.Cell(0, 10, e + ": " + n)
			pdf.Ln(8)
		}

		filename := "Reporte_Asistencia.pdf"
		pdf.OutputFileAndClose(filename)

		m := gomail.NewMessage()
		m.SetHeader("From", "kar.nunez34@unach.mx")
		m.SetHeader("To", "luis.gutierrez@unach.mx")
		m.SetHeader("Subject", "Reporte: " + nombreClase)
		m.Attach(filename)

		d := gomail.NewDialer("smtp.gmail.com", 587, "kar.nunez34@unach.mx", "feik wscy jlze pxqi")
		d.DialAndSend(m)

		return c.JSON(fiber.Map{"mensaje": "Reporte enviado con éxito"})
	})

	log.Fatal(app.Listen("0.0.0.0:3000"))
}