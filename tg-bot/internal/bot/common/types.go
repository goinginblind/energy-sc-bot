package common

import (
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/storage"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/client"
	"github.com/goinginblind/energy-sc-bot/tg-bot/ragpb"
	"go.uber.org/zap"
)

// HandlerEnv holds all the shared dependencies for the bot handlers.
type HandlerEnv struct {
	API        BotAPI
	Store      storage.Storage
	RAGClient  ragpb.RAGServiceClient
	DataClient client.DataClientInterface
	Logger     *zap.Logger
}