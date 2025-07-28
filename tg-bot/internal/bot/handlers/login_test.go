package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/common"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/bot/states"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/client"
	"github.com/goinginblind/energy-sc-bot/tg-bot/internal/logger"
	"github.com/stretchr/testify/require"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MockDataServiceClient is a mock implementation of the DataServiceClient for testing.
type MockDataServiceClient struct {
	GetUserByPhoneFunc func(phone string) (*client.User, error)
}

func (m *MockDataServiceClient) GetUserByPhone(phone string) (*client.User, error) {
	if m.GetUserByPhoneFunc != nil {
		return m.GetUserByPhoneFunc(phone)
	}
	return &client.User{ID: 1, Email: sql.NullString{String: "test@test.com", Valid: true}}, nil
}

func (m *MockDataServiceClient) GetBillsByUserID(userID int64) ([]client.Bill, error) {
	return nil, nil
}

func (m *MockDataServiceClient) Close() error {
	return nil
}

func TestHandleAwaitingLoginInput(t *testing.T) {
	logger.InitLogger()
	log := logger.L()

	ctx := context.Background()
	mockStorage := NewMockStorage()
	mockAPI := &MockBotAPI{}
	env := &common.HandlerEnv{
		API:    mockAPI,
		Store:  mockStorage,
		Logger: log,
	}

	chatID := int64(123)
	userID := int64(456)
	phone := "1234567890"

	HandleAwaitingLoginInput(ctx, env, chatID, userID, phone)

	// Check state and stored data
	userState, _ := mockStorage.GetUserState(ctx, userID)
	require.Equal(t, states.StateAwaitingOTP, userState)
	storedPhone, _ := mockStorage.GetUserData(ctx, userID, "login_identifier")
	require.Equal(t, phone, storedPhone)
	otp, _ := mockStorage.GetUserData(ctx, userID, "otp")
	require.NotEmpty(t, otp)

	// Check that a message was sent
	require.Len(t, mockAPI.SentMessages, 1)
	msg := mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
	require.Contains(t, msg.Text, "A confirmation code has been sent")
}

func TestHandleAwaitingOTP(t *testing.T) {
	logger.InitLogger()
	log := logger.L()

	ctx := context.Background()

	correctOTP := "123456"
	incorrectOTP := "654321"

	chatID := int64(123)
	userID := int64(456)

	t.Run("Correct OTP", func(t *testing.T) {
		mockStorage := NewMockStorage()
		mockAPI := &MockBotAPI{}
		mockDataClient := &MockDataServiceClient{}
		env := &common.HandlerEnv{
			API:        mockAPI,
			Store:      mockStorage,
			DataClient: mockDataClient,
			Logger:     log,
		}

		// Setup initial state
		mockStorage.SetUserState(ctx, userID, states.StateAwaitingOTP)
		mockStorage.SetUserData(ctx, userID, "otp", correctOTP)
		mockStorage.SetUserData(ctx, userID, "login_identifier", "1234567890")

		mockDataClient.GetUserByPhoneFunc = func(phone string) (*client.User, error) {
			return &client.User{ID: 1, Email: sql.NullString{String: "test@test.com", Valid: true}}, nil
		}
		HandleAwaitingOTP(ctx, env, chatID, userID, correctOTP)

		userState, _ := mockStorage.GetUserState(ctx, userID)
		require.Equal(t, states.StateLoggedIn, userState)
		loggedIn, _ := mockStorage.GetUserData(ctx, userID, "logged_in")
		require.Equal(t, "true", loggedIn)
		require.Len(t, mockAPI.SentMessages, 1)
		msg := mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
		require.Contains(t, msg.Text, "Login successful")
	})

	t.Run("Incorrect OTP", func(t *testing.T) {
		mockStorage := NewMockStorage()
		mockAPI := &MockBotAPI{}
		mockDataClient := &MockDataServiceClient{}
		env := &common.HandlerEnv{
			API:        mockAPI,
			Store:      mockStorage,
			DataClient: mockDataClient,
			Logger:     log,
		}

		// Setup initial state
		mockStorage.SetUserState(ctx, userID, states.StateAwaitingOTP)
		mockStorage.SetUserData(ctx, userID, "otp", correctOTP) // Set correct OTP for comparison
		mockStorage.SetUserData(ctx, userID, "login_identifier", "1234567890")

		HandleAwaitingOTP(ctx, env, chatID, userID, incorrectOTP)

		userState, _ := mockStorage.GetUserState(ctx, userID)
		require.Equal(t, states.StateAwaitingOTP, userState) // State should not change
		require.Len(t, mockAPI.SentMessages, 1)
		msg := mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
		require.Contains(t, msg.Text, "Incorrect code")
	})

	t.Run("User not found", func(t *testing.T) {
		mockStorage := NewMockStorage()
		mockAPI := &MockBotAPI{}
		mockDataClient := &MockDataServiceClient{}
		env := &common.HandlerEnv{
			API:        mockAPI,
			Store:      mockStorage,
			DataClient: mockDataClient,
			Logger:     log,
		}

		// Setup initial state
		mockStorage.SetUserState(ctx, userID, states.StateAwaitingOTP)
		mockStorage.SetUserData(ctx, userID, "otp", correctOTP) // Set correct OTP for comparison
		mockStorage.SetUserData(ctx, userID, "login_identifier", "1234567890")

		mockDataClient.GetUserByPhoneFunc = func(phone string) (*client.User, error) {
			return nil, fmt.Errorf("user not found")
		}
		HandleAwaitingOTP(ctx, env, chatID, userID, correctOTP)

		userState, _ := mockStorage.GetUserState(ctx, userID)
		require.Equal(t, states.StateStart, userState)
		require.Len(t, mockAPI.SentMessages, 1)
		msg := mockAPI.SentMessages[0].(tgbotapi.MessageConfig)
		require.Contains(t, msg.Text, "Could not find your account")
	})
}