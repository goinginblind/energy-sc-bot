package handlers

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/logger"
	"github.com/goinginblind/energy-sc-bot/tg-bot/ragpb"
	"github.com/stretchr/testify/require"
)

// MockBotAPI is a mock implementation of the tgbotapi.BotAPI for testing.
type MockBotAPI struct {
	SentMessages []tgbotapi.Chattable
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SentMessages = append(m.SentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return nil, nil
}

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return nil
}

func (m *MockBotAPI) GetMe() (tgbotapi.User, error) {
	return tgbotapi.User{UserName: "testbot"}, nil
}

// MockRAGClient is a mock implementation of the RAGClient for testing.
type MockRAGClient struct {
	ragpb.RAGServiceClient
	ClassifyQueryFunc    func(ctx context.Context, in *ragpb.ClassifyRequest, opts ...grpc.CallOption) (*ragpb.ClassifyResponse, error)
	GetAnswerToQueryFunc func(ctx context.Context, in *ragpb.AnswerRequest, opts ...grpc.CallOption) (*ragpb.AnswerResponse, error)
}

func (m *MockRAGClient) ClassifyQuery(ctx context.Context, in *ragpb.ClassifyRequest, opts ...grpc.CallOption) (*ragpb.ClassifyResponse, error) {
	if m.ClassifyQueryFunc != nil {
		return m.ClassifyQueryFunc(ctx, in, opts...)
	}
	return &ragpb.ClassifyResponse{Label: "general"}, nil
}

func (m *MockRAGClient) GetAnswerToQuery(ctx context.Context, in *ragpb.AnswerRequest, opts ...grpc.CallOption) (*ragpb.AnswerResponse, error) {
	if m.GetAnswerToQueryFunc != nil {
		return m.GetAnswerToQueryFunc(ctx, in, opts...)
	}
	return &ragpb.AnswerResponse{Answer: "Mocked RAG Answer"}, nil
}

func (m *MockRAGClient) Close() error {
	return nil
}

func TestHandleStartState(t *testing.T) {
	logger.InitLogger()
	log := logger.L()

	ctx := context.Background()
	mockStorage := NewMockStorage()
	mockAPI := &MockBotAPI{}
	mockRAGClient := &MockRAGClient{}

	env := &common.HandlerEnv{
		API:       mockAPI,
		Store:     mockStorage,
		RAGClient: mockRAGClient,
		Logger:    log,
	}

	chatID := int64(123)
	userID := int64(456)

	// Test case 1: User sends "/login"
	HandleStartState(ctx, env, chatID, userID, "/login")
	userState, _ := mockStorage.GetUserState(ctx, userID)
	require.Equal(t, states.StateAwaitingLoginInput, userState)
	require.Len(t, mockAPI.SentMessages, 1)
	msg := mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
	require.Equal(t, "Please enter your phone or email to log in.", msg.Text)

	// Test case 2: User sends "🔎 General Inquiry"
	mockAPI.SentMessages = nil // Reset sent messages
	HandleStartState(ctx, env, chatID, userID, "🔎 General Inquiry")
	user_state, _ := mockStorage.GetUserState(ctx, userID)
	require.Equal(t, states.StateGeneralInquiry, user_state)
	require.Len(t, mockAPI.SentMessages, 1)
	msg = mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
	require.Equal(t, "You are in general inquiry mode. Just write your question. To exit, use the button below or the /start command.", msg.Text)

	// Test case 3: User sends "/start"
	mockAPI.SentMessages = nil // Reset sent messages
	HandleStartState(ctx, env, chatID, userID, "/start")
	require.Len(t, mockAPI.SentMessages, 1)
	msg = mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
	require.Equal(t, "Hello! I am your virtual assistant. How can I help you?", msg.Text)
	require.NotNil(t, msg.ReplyMarkup)

	// Test case 4: User sends an arbitrary message (default case)
	mockAPI.SentMessages = nil // Reset sent messages
	mockStorage.SetUserState(ctx, userID, states.StateStart) // Ensure state is reset

	expectedRAGAnswer := "This is a mocked answer from RAG."
	mockRAGClient.GetAnswerToQueryFunc = func(ctx context.Context, in *ragpb.AnswerRequest, opts ...grpc.CallOption) (*ragpb.AnswerResponse, error) {
		return &ragpb.AnswerResponse{Answer: expectedRAGAnswer}, nil
	}

	HandleStartState(ctx, env, chatID, userID, "What is my latest bill?")
	userState, _ = mockStorage.GetUserState(ctx, userID)
	require.Equal(t, states.StateGeneralInquiry, userState)
	require.Len(t, mockAPI.SentMessages, 1)
	msg = mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
	require.Equal(t, expectedRAGAnswer, msg.Text)
	require.NotNil(t, msg.ReplyMarkup)
}