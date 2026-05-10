from fastapi import FastAPI
from pydantic import BaseModel
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

app = FastAPI(title="Microservicio de Notificaciones UNACH")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

class Registro(BaseModel):
    alumno_id: str
    materia: str

@app.post("/notificar")
async def notificar(registro: Registro):
    # Aquí puedes añadir lógica extra, como guardar en un log o enviar un webhook
    print(f"📢 [Python] Notificación recibida: Alumno {registro.alumno_id} entró a la clase {registro.materia}")
    return {"status": "ok", "received": True}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)