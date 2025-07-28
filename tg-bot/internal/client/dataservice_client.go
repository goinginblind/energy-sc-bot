package client

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// User defines the local struct for user data, decoupling it from the data service's internal db package.
type User struct {
	ID         int64          `json:"id"`
	TelegramID int64          `json:"telegram_id"`
	Phone      sql.NullString `json:"phone"`
	Email      sql.NullString `json:"email"`
	CreatedAt  time.Time      `json:"created_at"`
}

// Bill defines the local struct for bill data.
type Bill struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	PdfUrl   string    `json:"pdf_url"`
	Amount   float64   `json:"amount"`
	Status   string    `json:"status"`
	IssuedAt time.Time `json:"issued_at"`
	DueDate  time.Time `json:"due_date"`
}

type DataServiceClient struct {
	BaseURL string
	Client  *http.Client
}

func NewDataServiceClient() *DataServiceClient {
	baseURL := os.Getenv("PERSONAL_DATA_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &DataServiceClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *DataServiceClient) GetUserByPhone(phone string) (*User, error) {
	resp, err := c.Client.Get(fmt.Sprintf("%s/user/phone/%s", c.BaseURL, phone))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: status code %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *DataServiceClient) GetBillsByUserID(userID int64) ([]Bill, error) {
	resp, err := c.Client.Get(fmt.Sprintf("%s/users/%d/bills", c.BaseURL, userID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get bills: status code %d", resp.StatusCode)
	}

	var bills []Bill
	if err := json.NewDecoder(resp.Body).Decode(&bills); err != nil {
		return nil, err
	}
	return bills, nil
}
