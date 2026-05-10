from fastapi import FastAPI
from pydantic import BaseModel
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

app = FastAPI(title="Microservicio de Notificaciones UNACH")

# --- CONFIGURACIÓN DE SEGURIDAD (CORS) ---
# Aquí permitimos que tanto tu IP de AWS como tu página de Vercel puedan hablar con Python
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://184.73.140.150:3000",      # Tu Backend de Go en AWS
        "https://sistema-unach.vercel.app", # Tu Frontend en Vercel
        "http://localhost:5173",            # Para tus pruebas locales
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class Registro(BaseModel):
    alumno_id: str
    materia: str

@app.post("/notificar")
async def notificar(registro: Registro):
    # Esta es la lógica que se dispara cuando Go le avisa a Python
    print(f"📢 [NOTIFICACIÓN] Alumno: {registro.alumno_id} | Materia: {registro.materia}")
    
    # Aquí es donde en el futuro podrías meter la IA de análisis facial 
    # o enviar mensajes por Telegram/WhatsApp.
    return {
        "status": "notificación procesada", 
        "alumno": registro.alumno_id,
        "materia": registro.materia
    }

if __name__ == "__main__":
    # Importante: host 0.0.0.0 permite que AWS reciba peticiones externas
    uvicorn.run(app, host="0.0.0.0", port=8001)