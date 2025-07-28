package client

// DataClientInterface defines the interface for the data service client.
type DataClientInterface interface {
	GetUserByPhone(phone string) (*User, error)
	GetBillsByUserID(userID int64) ([]Bill, error)
}

// Ensure that DataServiceClient implements DataClientInterface
var _ DataClientInterface = (*DataServiceClient)(nil)
