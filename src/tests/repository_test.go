package main_test

import (
    "context"
    "errors"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go_rest_wallets/app"
)

// ==============================
// Моки
// ==============================

type MockDB struct {
    mock.Mock
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
    allArgs := append([]any{ctx, sql}, args...)
    argsOut := m.Called(allArgs...)
    return argsOut.Get(0).(pgx.Row)
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
    allArgs := append([]any{ctx, sql}, args...)
    argsOut := m.Called(allArgs...)
    return argsOut.Get(0).(pgconn.CommandTag), argsOut.Error(1)
}

type mockRow struct {
    ScanFunc func(...any) error
}

func (r *mockRow) Scan(dest ...any) error {
    return r.ScanFunc(dest...)
}

// ==============================
// Хелперы
// ==============================

// makeMockRow — создаёт мокированную строку с id и amount
func makeMockRow(id string, amount float64) *mockRow {
    return &mockRow{
        ScanFunc: func(dest ...any) error {
            if len(dest) >= 2 {
                *(dest[0].(*string)) = id
                *(dest[1].(*float64)) = amount
            }
            return nil
        },
    }
}

// makeErrorMockRow — создаёт строку, которая возвращает ошибку при сканировании
func makeErrorMockRow(err error) *mockRow {
    return &mockRow{
        ScanFunc: func(dest ...any) error {
            return err
        },
    }
}

// makeCommandTag — оборачивает строку в pgconn.CommandTag
func makeCommandTag(s string) pgconn.CommandTag {
    return pgconn.NewCommandTag(s)
}

// ==============================
// Тесты
// ==============================

func Test_GetWalletById_Success(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "1").
        Return(makeMockRow("1", 100.0))

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.GetWalletById("1")

    assert.NoError(t, err)
    assert.Equal(t, &app.Wallet{Id: "1", Amount: 100.0}, wallet)
    mockDB.AssertExpectations(t)
}

func Test_GetWalletById_NotFound(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "not-found").
        Return(makeErrorMockRow(pgx.ErrNoRows))

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.GetWalletById("not-found")

    assert.ErrorIs(t, err, app.ErrWalletNotFound)
    assert.Nil(t, wallet)
    mockDB.AssertExpectations(t)
}

func Test_GetWalletById_DBError(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "broken").
        Return(makeErrorMockRow(errors.New("some DB error")))

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.GetWalletById("broken")

    assert.Error(t, err)
    assert.Nil(t, wallet)
    assert.Contains(t, err.Error(), "some DB error")
    mockDB.AssertExpectations(t)
}

func Test_Deposit_ExistingWallet(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "1").
        Return(makeMockRow("1", 100.0))

    mockDB.On("Exec", mock.Anything, "update wallets set amount = $1 where id = $2", 150.0, "1").
        Return(makeCommandTag("UPDATE 1"), nil)

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.Deposit("1", 50.0)

    assert.NoError(t, err)
    assert.Equal(t, &app.Wallet{Id: "1", Amount: 150.0}, wallet)
    mockDB.AssertExpectations(t)
}

func Test_Deposit_NewWallet(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "insert into wallets (amount) values ($1) returning id", 100.0).
        Return(&mockRow{
            ScanFunc: func(dest ...interface{}) error {
                *(dest[0].(*string)) = "new-id"
                return nil
            },
        })

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.Deposit("", 100.0)

    assert.NoError(t, err)
    assert.Equal(t, &app.Wallet{Id: "new-id", Amount: 100.0}, wallet)
    mockDB.AssertExpectations(t)
}

func Test_Withdraw_InsufficientFunds(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "1").
        Return(makeMockRow("1", 50.0))

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.Withdraw("1", 100.0)

    assert.ErrorIs(t, err, &app.ValidationError{})
    assert.Nil(t, wallet)
    mockDB.AssertExpectations(t)
}

func Test_Withdraw_Success(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "1").
        Return(makeMockRow("1", 100.0))

    mockDB.On("Exec", mock.Anything, "update wallets set amount = $1 where id = $2", 50.0, "1").
        Return(makeCommandTag("UPDATE 1"), nil)

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.Withdraw("1", 50.0)

    assert.NoError(t, err)
    assert.Equal(t, &app.Wallet{Id: "1", Amount: 50.0}, wallet)
    mockDB.AssertExpectations(t)
}

func Test_Withdraw_QueryError(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "1").
        Return(makeErrorMockRow(errors.New("db query failed")))

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.Withdraw("1", 50.0)

    assert.Error(t, err)
    assert.Nil(t, wallet)
    assert.Contains(t, err.Error(), "db query failed")
    mockDB.AssertExpectations(t)
}

func Test_Withdraw_UpdateError(t *testing.T) {
    mockDB := new(MockDB)

    mockDB.On("QueryRow", mock.Anything, "select id, amount from wallets where id = $1", "1").
        Return(makeMockRow("1", 100.0))

    mockDB.On("Exec", mock.Anything, "update wallets set amount = $1 where id = $2", 50.0, "1").
        Return(makeCommandTag("UPDATE 0"), errors.New("update failed"))

    repo := &app.WalletsRepository{Conn: mockDB}
    wallet, err := repo.Withdraw("1", 50.0)

    assert.Error(t, err)
    assert.Nil(t, wallet)
    assert.Contains(t, err.Error(), "update failed")
    mockDB.AssertExpectations(t)
}
