package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wallet-service/models"
	"wallet-service/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	cfg := &config.Config{
		PostgresHost:     "localhost",
		PostgresPort:     5432,
		PostgresUser:     "walletuser",
		PostgresPassword: "walletpass",
		PostgresDB:       "walletdb_test",
		MaxDBConnections: 5,
	}

	pool, err := db.NewPostgresPool(cfg)
	require.NoError(t, err)

	// Clean up before tests
	_, err = pool.Exec(context.Background(), "DROP TABLE IF EXISTS transactions, wallets")
	require.NoError(t, err)

	// Apply migrations
	migrateDB(t, pool)

	return pool
}

func migrateDB(t *testing.T, pool *pgxpool.Pool) {
	// Read migration file
	migrationSQL, err := os.ReadFile("db/migrations/000001_init_schema.up.sql")
	require.NoError(t, err)

	// Execute migration
	_, err = pool.Exec(context.Background(), string(migrationSQL))
	require.NoError(t, err)
}

func TestWalletService(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := services.NewWalletService(pool)
	ctx := context.Background()

	t.Run("Create wallet on first deposit", func(t *testing.T) {
		walletID := uuid.New()
		op := models.WalletOperation{
			WalletID:      walletID,
			OperationType: models.Deposit,
			Amount:       1000,
		}

		err := service.ProcessOperation(ctx, op)
		assert.NoError(t, err)

		balance, err := service.GetBalance(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), balance)
	})

	t.Run("Withdraw from wallet", func(t *testing.T) {
		walletID := uuid.New()
		
		// Initial deposit
		err := service.ProcessOperation(ctx, models.WalletOperation{
			WalletID:      walletID,
			OperationType: models.Deposit,
			Amount:       2000,
		})
		assert.NoError(t, err)

		// Withdraw
		err = service.ProcessOperation(ctx, models.WalletOperation{
			WalletID:      walletID,
			OperationType: models.Withdraw,
			Amount:       500,
		})
		assert.NoError(t, err)

		balance, err := service.GetBalance(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1500), balance)
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		walletID := uuid.New()
		
		// Initial deposit
		err := service.ProcessOperation(ctx, models.WalletOperation{
			WalletID:      walletID,
			OperationType: models.Deposit,
			Amount:       100,
		})
		assert.NoError(t, err)

		// Attempt to withdraw more than balance
		err = service.ProcessOperation(ctx, models.WalletOperation{
			WalletID:      walletID,
			OperationType: models.Withdraw,
			Amount:       200,
		})
		assert.Equal(t, services.ErrInsufficientFunds, err)

		balance, err := service.GetBalance(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(100), balance)
	})

	t.Run("Concurrent operations", func(t *testing.T) {
		walletID := uuid.New()
		const operations = 100
		const amount = 10

		// Initial deposit
		err := service.ProcessOperation(ctx, models.WalletOperation{
			WalletID:      walletID,
			OperationType: models.Deposit,
			Amount:       operations * amount,
		})
		assert.NoError(t, err)

		// Run concurrent withdrawals
		errors := make(chan error, operations)
		for i := 0; i < operations; i++ {
			go func() {
				err := service.ProcessOperation(ctx, models.WalletOperation{
					WalletID:      walletID,
					OperationType: models.Withdraw,
					Amount:       amount,
				})
				errors <- err
			}()
		}

		// Wait for all operations to complete
		for i := 0; i < operations; i++ {
			err := <-errors
			assert.NoError(t, err)
		}

		balance, err := service.GetBalance(ctx, walletID)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), balance)
	})
}

func TestWalletAPI(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	service := services.NewWalletService(pool)
	handler := api.NewWalletHandler(service)

	t.Run("POST /api/v1/wallet - deposit", func(t *testing.T) {
		walletID := uuid.New()
		payload := `{"walletId": "` + walletID.String() + `", "operationType": "DEPOSIT", "amount": 1000}`
		req := httptest.NewRequest("POST", "/api/v1/wallet", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleOperation(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Check balance
		req = httptest.NewRequest("GET", "/api/v1/wallets/"+walletID.String(), nil)
		w = httptest.NewRecorder()

		handler.HandleGetBalance(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response models.WalletBalanceResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000), response.Balance)
	})

	t.Run("GET /api/v1/wallets/{walletId} - not found", func(t *testing.T) {
		walletID := uuid.New()
		req := httptest.NewRequest("GET", "/api/v1/wallets/"+walletID.String(), nil)
		w := httptest.NewRecorder()

		handler.HandleGetBalance(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
