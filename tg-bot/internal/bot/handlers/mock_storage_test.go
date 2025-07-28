package handlers

import (
	"context"
	"strings"
)

// MockStorage is a mock implementation of the storage.Storage interface for testing.
type MockStorage struct {
	UserState map[int64]string
	UserData  map[int64]map[string]string
	History   map[int64][]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		UserState: make(map[int64]string),
		UserData:  make(map[int64]map[string]string),
		History:   make(map[int64][]string),
	}
}

func (m *MockStorage) GetUserState(ctx context.Context, userID int64) (string, error) {
	return m.UserState[userID], nil
}

func (m *MockStorage) SetUserState(ctx context.Context, userID int64, state string) error {
	m.UserState[userID] = state
	return nil
}

func (m *MockStorage) GetUserData(ctx context.Context, userID int64, key string) (string, error) {
	if m.UserData[userID] == nil {
		return "", nil
	}
	return m.UserData[userID][key], nil
}

func (m *MockStorage) SetUserData(ctx context.Context, userID int64, key, value string) error {
	if m.UserData[userID] == nil {
		m.UserData[userID] = make(map[string]string)
	}
	m.UserData[userID][key] = value
	return nil
}

func (m *MockStorage) ClearUserData(ctx context.Context, userID int64) error {
	delete(m.UserData, userID)
	return nil
}

func (m *MockStorage) SaveMessage(ctx context.Context, userID int64, message string) error {
	m.History[userID] = append(m.History[userID], message)
	return nil
}

func (m *MockStorage) GetHistory(ctx context.Context, userID int64, limit int) (string, error) {
	history := m.History[userID]
	if len(history) > limit {
		history = history[len(history)-limit:]
	}
	return strings.Join(history, "\n"), nil
}

func (m *MockStorage) Close() error {
	return nil
}
