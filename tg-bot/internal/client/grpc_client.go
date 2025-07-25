package client

import (
	"log"

	"github.com/goinginblind/energy-sc-bot/tg-bot/ragpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Новый клиент, аргумент - адрес gRPC сервиса
func New(grpcAddr string) ragpb.RAGServiceClient {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}

	client := ragpb.NewRAGServiceClient(conn)
	return client
}
