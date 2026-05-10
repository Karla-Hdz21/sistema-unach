from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(title="Servicio de Validación Visual - LIDTS")

# Configuración de CORS para tu frontend de React
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173"],
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
def read_root():
    return {"status": "online", "microservice": "Visual Validator 6N"}

@app.post("/validate-image")
async def validate_image(data: dict):
    # Aquí iría la lógica de análisis de imagen
    return {"valid": True, "confidence": 0.98, "message": "Imagen verificada correctamente"}