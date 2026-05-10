FROM python:3.10-slim
WORKDIR /app
COPY . .
RUN pip install fastapi uvicorn pydantic
EXPOSE 8001
CMD ["python", "main.py"]