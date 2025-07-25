FROM python:3.12-slim

WORKDIR /app

# Install system dependencies (for FAISS & Unstructured)
RUN apt-get update && apt-get install -y \
    build-essential \
    libopenblas-dev \
    libgomp1 \
    git \
 && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY rag/requirements.txt .
RUN pip install --upgrade pip && pip install -r requirements.txt \
    && pip install flask grpcio grpcio-tools

# Copy entire RAG codebase
COPY rag/ ./
COPY rag-service/ ./rag-service/

EXPOSE 50051
CMD ["python", "rag-service/main.py"]
