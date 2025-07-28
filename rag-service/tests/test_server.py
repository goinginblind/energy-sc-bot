import grpc
import pytest
from concurrent import futures
from unittest.mock import patch, MagicMock

from server.server import RAGServicer
import server.ragpb.rag_pb2 as rag_pb2
import server.ragpb.rag_pb2_grpc as rag_pb2_grpc

@pytest.fixture(scope="module")
def grpc_server():
    # Mock the make_rag function to avoid real model loading
    with patch('server.server.make_rag') as mock_make_rag:
        mock_rag_model = MagicMock()
        mock_rag_model.embedding_model.api_key = "fake_key"
        mock_make_rag.return_value = mock_rag_model

        server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
        rag_pb2_grpc.add_RAGServiceServicer_to_server(
            RAGServicer(key="fake_key", db_path="fake_path", prompts_path="fake_path"), server
        )
        server.add_insecure_port('[::]:50051')
        server.start()
        yield server
        server.stop(0)

def test_classify_query(grpc_server):
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = rag_pb2_grpc.RAGServiceStub(channel)
        
        # Mock the classifier function
        with patch('server.server.classifier') as mock_classifier:
            mock_classifier.return_value = "1. Жалоба пользователя"
            
            response = stub.ClassifyQuery(rag_pb2.ClassifyRequest(query="test query"))
            assert response.label == "1. Жалоба пользователя"

def test_get_answer_to_query(grpc_server):
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = rag_pb2_grpc.RAGServiceStub(channel)

        # Mock the make_prompt and get_answer_to_query functions
        with patch('server.server.make_prompt') as mock_make_prompt, \
             patch('server.server.get_answer_to_query') as mock_get_answer:
            
            mock_make_prompt.return_value = "test prompt"
            mock_get_answer.return_value = "test answer"

            response = stub.GetAnswerToQuery(rag_pb2.AnswerRequest(query="test query"))
            assert response.answer == "test answer"

def test_human_query_to_prompt(grpc_server):
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = rag_pb2_grpc.RAGServiceStub(channel)

        # Mock the human_query_to_gpt_prompt function
        with patch('server.server.human_query_to_gpt_prompt') as mock_human_query:
            mock_human_query.return_value = "rephrased query"

            response = stub.HumanQueryToPrompt(rag_pb2.HumanQueryRequest(query="test query"))
            assert response.rephrased_query == "rephrased query"

