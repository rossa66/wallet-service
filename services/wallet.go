package services

import (
	"context"
	"errors"
	"fmt"
	"wallet-service/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrWalletNotFound    = errors.New("wallet not found")
)

type WalletService struct {
	db *pgxpool.Pool
}

func NewWalletService(db *pgxpool.Pool) *WalletService {
	return &WalletService{db: db}
}

func (s *WalletService) ProcessOperation(ctx context.Context, op models.WalletOperation) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if wallet exists
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)", op.WalletID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check wallet existence: %w", err)
	}

	if !exists {
		// Create wallet if it doesn't exist
		_, err = tx.Exec(ctx, "INSERT INTO wallets (id, balance) VALUES ($1, $2)", op.WalletID, 0)
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}
	}

	var balance int64
	err = tx.QueryRow(ctx, "SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", op.WalletID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("failed to get wallet balance: %w", err)
	}

	newBalance := balance
	switch op.OperationType {
	case models.Deposit:
		newBalance += op.Amount
	case models.Withdraw:
		if balance < op.Amount {
			return ErrInsufficientFunds
		}
		newBalance -= op.Amount
	default:
		return fmt.Errorf("invalid operation type: %s", op.OperationType)
	}

	// Update wallet balance
	_, err = tx.Exec(ctx, "UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2", newBalance, op.WalletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Record transaction
	_, err = tx.Exec(ctx, 
		"INSERT INTO transactions (wallet_id, operation_type, amount) VALUES ($1, $2, $3)",
		op.WalletID, op.OperationType, op.Amount)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var balance int64
	err := s.db.QueryRow(ctx, "SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, ErrWalletNotFound
		}
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}
	return balance, nil
}
