package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Wallet struct {
	Id     string  `json:"id"`
	Amount float64 `json:"amount"`
}

func (w *Wallet) ToJson() []byte {
	jsonData, err := json.Marshal(w)

	if err != nil {
		return []byte("{}")
	}

	return jsonData
}

var ErrWalletNotFound = errors.New("wallet not found")

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation error on field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Is(target error) bool {
    _, ok := target.(*ValidationError)
    return ok
}

type WalletsRepositoryInterface interface {
	GetWalletById(id string) (*Wallet, error)
	Deposit(id string, amount float64) (*Wallet, error)
	Withdraw(id string, amount float64) (*Wallet, error)
}

type DbConnectionInterface interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type WalletsRepository struct {
	Conn DbConnectionInterface
}

// Получить кошелек по ID
func (r *WalletsRepository) GetWalletById(id string) (*Wallet, error) {
	if id == "" {
		return nil, &ValidationError{Field: "id", Message: "cannot be empty"}
	}

	row := r.Conn.QueryRow(context.Background(), "select id, amount from wallets where id = $1", id)

	var wallet Wallet
	err := row.Scan(&wallet.Id, &wallet.Amount)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletNotFound
	}

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// Пополнение
func (r *WalletsRepository) Deposit(id string, amount float64) (*Wallet, error) {
	return r.walletTransaction(id, amount, true)
}

// Снятие
func (r *WalletsRepository) Withdraw(id string, amount float64) (*Wallet, error) {
	return r.walletTransaction(id, amount, false)
}

// todo: сделать чтобы можно было не указывать id, а он создавался автоматически
// и протестировать производительность этого метода

func (r *WalletsRepository) walletTransaction(id string, amount float64, add bool) (*Wallet, error) {
	if id == "" {
		if add {
			var newId string
			err := r.Conn.QueryRow(context.Background(), "insert into wallets (amount) values ($1) returning id", amount).Scan(&newId)

			if err != nil {
				return nil, err
			}

			return &Wallet{Id: newId, Amount: amount}, nil
		}

		return nil, &ValidationError{Field: "id", Message: "cannot be empty"}
	}

	if amount < 0 {
		return nil, &ValidationError{Field: "amount", Message: "must be positive"}
	}

	wallet, err := r.GetWalletById(id)

	if err == ErrWalletNotFound {
		if add {
			msg := fmt.Errorf("failed to create new wallet with id %s", id)
			err := r.execQuery("insert into wallets (id, amount) values ($1, $2)", msg, id, amount)

			if err != nil {
				return nil, err
			}

			return &Wallet{Id: id, Amount: amount}, nil
		}

		return nil, err
	}

	if err != nil {
		return nil, err
	}

	if add {
		wallet.Amount += amount
	} else {
		if wallet.Amount < amount {
			return nil, &ValidationError{Field: "amount", Message: "cannot be greater than current wallet amount"}
		}

		wallet.Amount -= amount
	}

	msg := fmt.Errorf("failed to update wallet with id %s", id)
	err = r.execQuery("update wallets set amount = $1 where id = $2", msg, wallet.Amount, id)

	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func (r *WalletsRepository) execQuery(query string, noRowsError error, args ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := r.Conn.Exec(ctx, query, args...)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return noRowsError
	}

	return nil
}
