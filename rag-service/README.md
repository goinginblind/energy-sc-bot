# RAG Service (Python Retrieval-Augmented Generation)

This service is a Python-based gRPC server that provides Retrieval-Augmented Generation (RAG) capabilities. It's designed to classify user queries, retrieve relevant information from a knowledge base, and generate answers using a Large Language Model (LLM).

## Technologies Used

*   **Python:** The primary language for the service.
*   **gRPC:** For high-performance communication with the Telegram bot.
*   **`transformers`:** (Likely) For LLM integration.
*   **`faiss-cpu` / `scikit-learn` / `numpy`:** (Likely) For vector storage and similarity search in the knowledge base.
*   **`pytest`:** For testing.

## Setup

### Environment Variables

Create a `.env` file in the `rag-service/` directory if any specific API keys or model paths are required. For basic operation, it might not need any.

```env
# Example (if using OpenAI API)
# OPENAI_API_KEY=your_openai_api_key

# Example (if using a local model or specific model path)
# MODEL_PATH=/path/to/your/model
```

### Knowledge Base

The RAG service relies on a knowledge base, typically stored as vector embeddings (e.g., `index.faiss`, `index.pkl` as seen in `rag/docs/`). This knowledge base needs to be pre-built or generated. The process for building or updating this knowledge base is not explicitly defined in the provided files but is a crucial part of the RAG system.

### Build and Run Locally

1.  **Navigate to the service directory:**
    ```bash
    cd rag-service
    ```
2.  **Install Python dependencies:**
    ```bash
    pip install -r requirements.txt
    ```
3.  **Run the application:**
    ```bash
    python main.py
    ```
    The gRPC server will start, typically listening on port `50051` (as configured in `docker-compose.yml`).

## Functionality

The RAG service exposes the following gRPC methods (defined in `proto/rag.proto`):

*   `ClassifyQuery`: Classifies the user's input query into predefined categories.
*   `GetAnswerToQuery`: Generates a comprehensive answer to the user's query, potentially using retrieved information from the knowledge base and an LLM.
*   `HumanQueryToPrompt`: (Likely) Transforms a human-readable query into a structured prompt for the LLM.

## Testing

To run unit and integration tests for the RAG service:

```bash
cd rag-service
pytest
```
