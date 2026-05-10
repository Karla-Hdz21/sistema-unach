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

// --- CAMBIO 1: La URL de Python debe usar la IP de AWS ---
func avisarAPython(alumnoID string, materia string) {
	url := "http://184.73.140.150:8001/notificar" 
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
	// --- CAMBIO 2: La conexión a la DB suele ser local dentro del servidor ---
	// Si tu base de datos corre en el mismo AWS, dejamos 'localhost' o usamos 'db' si usas Docker
	connStr := "postgresql://postgres:unah2026@localhost:5432/sistema_unach?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil { log.Fatal(err) }

	app := fiber.New()

	// --- CAMBIO 3: Permitir que Vercel se conecte (CORS) ---
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://sistema-unach.vercel.app, http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Post("/asistencia", func(c *fiber.Ctx) error {
		var req AsistenciaRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"mensaje": "Error en los datos"})
		}
		
		var nombreClase string
		db.QueryRow("SELECT nombre FROM clases WHERE id = $1", req.ClaseID).Scan(&nombreClase)
		
		if !esHorarioPermitido(nombreClase) {
			return c.Status(403).JSON(fiber.Map{"mensaje": "Acceso Denegado: Fuera de horario"})
		}

		var existe int
		db.QueryRow("SELECT COUNT(*) FROM asistencias WHERE alumno_id=$1 AND clase_id=$2 AND fecha_hora::date = CURRENT_DATE", req.AlumnoID, req.ClaseID).Scan(&existe)
		if existe > 0 {
			return c.Status(400).JSON(fiber.Map{"mensaje": "Ya tienes asistencia hoy"})
		}

		var nombreAlum string
		db.QueryRow("SELECT nombre FROM alumnos WHERE id = $1", req.AlumnoID).Scan(&nombreAlum)

		db.Exec("INSERT INTO asistencias (alumno_id, clase_id, fecha_hora) VALUES ($1, $2, NOW())", req.AlumnoID, req.ClaseID)
		
		go avisarAPython(req.AlumnoID, nombreClase)
		return c.JSON(fiber.Map{"mensaje": "¡Bienvenido, " + nombreAlum + "!"})
	})

	app.Post("/cerrar-clase", func(c *fiber.Ctx) error {
		var req struct { ClaseID int `json:"clase_id"` }
		c.BodyParser(&req)
		
		var nombreClase string
		db.QueryRow("SELECT nombre FROM clases WHERE id = $1", req.ClaseID).Scan(&nombreClase)

		rows, _ := db.Query(`
			SELECT a.nombre, 
			CASE WHEN asis.id IS NULL THEN 'Faltó' ELSE 'Asistió' END as estado
			FROM alumnos a
			LEFT JOIN asistencias asis ON a.id = asis.alumno_id 
			AND asis.clase_id = $1 
			AND asis.fecha_hora::date = CURRENT_DATE`, req.ClaseID)
		
		var asistentes, faltantes []string
		for rows.Next() {
			var n, e string
			rows.Scan(&n, &e)
			if e == "Asistió" { asistentes = append(asistentes, n) } else { faltantes = append(faltantes, n) }
		}

		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, "REPORTE 6N LIDTS - "+nombreClase)
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 12)
		for _, nom := range asistentes { pdf.SetTextColor(0, 150, 0); pdf.Cell(0, 10, "PRESENTE: "+nom); pdf.Ln(8) }
		for _, nom := range faltantes { pdf.SetTextColor(200, 0, 0); pdf.Cell(0, 10, "FALTA: "+nom); pdf.Ln(8) }

		filename := "Reporte_Asistencia.pdf"
		pdf.OutputFileAndClose(filename)

		m := gomail.NewMessage()
		m.SetHeader("From", "kar.nunez34@unach.mx")
		m.SetHeader("To", "luis.gutierrez@unach.mx") 
		m.SetHeader("Subject", "REPORTE FINAL: "+nombreClase)
		m.SetBody("text/plain", "Se adjunta el reporte oficial del grupo 6N LIDTS.")
		m.Attach(filename)

		d := gomail.NewDialer("smtp.gmail.com", 587, "kar.nunez34@unach.mx", "feik wscy jlze pxqi")
		if err := d.DialAndSend(m); err != nil {
			return c.Status(500).JSON(fiber.Map{"mensaje": "Error al enviar correo"})
		}

		return c.JSON(fiber.Map{"mensaje": "Reporte enviado con éxito"})
	})

	app.Get("/clases", func(c *fiber.Ctx) error {
		rows, _ := db.Query("SELECT id, nombre FROM clases")
		var clases []fiber.Map
		for rows.Next() {
			var id int; var nom string
			rows.Scan(&id, &nom)
			clases = append(clases, fiber.Map{"id": id, "nombre": nom})
		}
		return c.JSON(clases)
	})

	// --- CAMBIO 4: Escuchar en todas las interfaces ---
	log.Fatal(app.Listen("0.0.0.0:3000"))
}