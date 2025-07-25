FROM python:3.11-slim

WORKDIR /app

# Установить системные зависимости для FAISS и Unstructured
RUN apt-get update && apt-get install -y \
    build-essential \
    libopenblas-dev \
    libgomp1 \
    git \
    && rm -rf /var/lib/apt/lists/*

# Копировать requirements и установить зависимости
COPY rag/requirements.txt ./requirements.txt
RUN pip install --upgrade pip && pip install -r requirements.txt && pip install flask grpcio grpcio-tools

# Копировать исходный код
COPY rag/ rag/
COPY rag-service/ rag-service/

# Открыть порт для gRPC
EXPOSE 50051

# Запускать 
CMD ["python", "rag-service/main.py"]
