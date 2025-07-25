import grpc
from concurrent import futures
import server.ragpb.rag_pb2 as rag_pb2
import server.ragpb.rag_pb2_grpc as rag_pb2_grpc

from rag.rag import make_rag, make_prompt, get_answer_to_query, classifier, human_query_to_gpt_prompt

class RAGServicer(rag_pb2_grpc.RAGServiceServicer):
    def __init__(self, key):
        self.rag_model = make_rag(key=key)

    def GetAnswerToQuery(self, request, context):
        prompt = make_prompt(face=request.face, history=request.history, acc_data=request.acc_data, label=request.label)
        answer = get_answer_to_query(request.query, prompt, self.rag_model)
        return rag_pb2.AnswerResponse(answer=answer)

    def ClassifyQuery(self, request, context):
        result = classifier(key=self.rag_model.embedding_model.api_key, query=request.query)
        return rag_pb2.ClassifyResponse(label=result)

    def HumanQueryToPrompt(self, request, context):
        result = human_query_to_gpt_prompt(key=self.rag_model.embedding_model.api_key, query=request.query)
        return rag_pb2.HumanQueryResponse(rephrased_query=result)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    rag_pb2_grpc.add_RAGServiceServicer_to_server(RAGServicer(key='YOUR_OPENAI_KEY'), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    print("RAG server running on port 50051")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
