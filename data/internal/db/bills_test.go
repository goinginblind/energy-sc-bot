package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createRandomBill(t *testing.T, user User) Bill {
	arg := CreateBillParams{
		UserID:    user.ID,
		Amount:    100.50,
		Status:    "unpaid",
		IssuedAt:  time.Now(),
		DueDate:   time.Now().AddDate(0, 1, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bill, err := testQueries.CreateBill(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, bill)

	require.Equal(t, arg.UserID, bill.UserID)
	require.Equal(t, arg.Amount, bill.Amount)
	require.Equal(t, arg.Status, bill.Status)
	require.WithinDuration(t, arg.IssuedAt, bill.IssuedAt, time.Second)
	require.WithinDuration(t, arg.DueDate, bill.DueDate, time.Second)

	require.NotZero(t, bill.ID)
	require.NotZero(t, bill.CreatedAt)
	require.NotZero(t, bill.UpdatedAt)

	return bill
}

func TestCreateBill(t *testing.T) {
	user := createRandomUser(t)
	createRandomBill(t, user)
}

func TestGetBill(t *testing.T) {
	user := createRandomUser(t)
	bill1 := createRandomBill(t, user)
	bill2, err := testQueries.GetBill(context.Background(), bill1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, bill2)

	require.Equal(t, bill1.ID, bill2.ID)
	require.Equal(t, bill1.UserID, bill2.UserID)
	require.Equal(t, bill1.Amount, bill2.Amount)
	require.Equal(t, bill1.Status, bill2.Status)
	require.WithinDuration(t, bill1.IssuedAt, bill2.IssuedAt, time.Second)
	require.WithinDuration(t, bill1.DueDate, bill2.DueDate, time.Second)
	require.WithinDuration(t, bill1.CreatedAt, bill2.CreatedAt, time.Second)
	require.WithinDuration(t, bill1.UpdatedAt, bill2.UpdatedAt, time.Second)
}

func TestGetBillsByUser(t *testing.T) {
	user := createRandomUser(t)
	for i := 0; i < 10; i++ {
		createRandomBill(t, user)
	}

	bills, err := testQueries.GetBillsByUserID(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, bills, 10)

	for _, bill := range bills {
		require.Equal(t, user.ID, bill.UserID)
	}
}

func TestUpdateBill(t *testing.T) {
	user := createRandomUser(t)
	bill1 := createRandomBill(t, user)

	arg := UpdateBillParams{
		ID:     bill1.ID,
		Status: "paid",
	}

	bill2, err := testQueries.UpdateBill(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, bill2)

	require.Equal(t, bill1.ID, bill2.ID)
	require.Equal(t, "paid", bill2.Status)
	require.NotEqual(t, bill1.UpdatedAt, bill2.UpdatedAt)
}
