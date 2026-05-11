from fastapi import FastAPI
from pydantic import BaseModel
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

app = FastAPI(title="Microservicio de Notificaciones UNACH")

# --- CONFIGURACIÓN DE SEGURIDAD (CORS) ---
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://54.173.32.242:3000",        # ACTUALIZADO: Tu nueva IP Elástica
        "https://sistema-unach.vercel.app", 
        "http://localhost:5173",            
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
    print(f"📢 [NOTIFICACIÓN] Alumno: {registro.alumno_id} | Materia: {registro.materia}")
    return {
        "status": "notificación procesada", 
        "alumno": registro.alumno_id,
        "materia": registro.materia
    }

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)