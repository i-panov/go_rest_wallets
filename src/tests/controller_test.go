package main_test

import (
	"encoding/json"
	"go_rest_wallets/app"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==============================
// Моки
// ==============================

type MockWalletRepository struct {
    mock.Mock
}

func (m *MockWalletRepository) GetWalletById(id string) (*app.Wallet, error) {
    args := m.Called(id)
    return args.Get(0).(*app.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Deposit(id string, amount float64) (*app.Wallet, error) {
    args := m.Called(id, amount)
    return args.Get(0).(*app.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Withdraw(id string, amount float64) (*app.Wallet, error) {
    args := m.Called(id, amount)
    return args.Get(0).(*app.Wallet), args.Error(1)
}

// ==============================
// Хелперы
// ==============================

func createFakeGetRequest(data map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", "", nil)
	return mux.SetURLVars(req, data)
}

func createFakePostRequest(data map[string]any) *http.Request {
	bytes, _ := json.Marshal(data)
	body := strings.NewReader(string(bytes))
	req, _ := http.NewRequest("POST", "", body)
	req.Header.Set("Content-Type", "application/json")
	return req
}

var (
    nilWallet *app.Wallet = nil
)

// ==============================
// Тесты
// ==============================

func Test_GetWallet_Handler_ReturnsWallet(t *testing.T) {
    mockRepo := new(MockWalletRepository)

    wallet := &app.Wallet{Id: "1", Amount: 100.0}

    mockRepo.On("GetWalletById", wallet.Id).Return(wallet, nil)

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

	req := createFakeGetRequest(map[string]string{
		"id": wallet.Id,
	})

	rr := httptest.NewRecorder()

    controller.GetWallet(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.JSONEq(t, `{"id":"1","amount":100}`, rr.Body.String())

    mockRepo.AssertExpectations(t)
}

func Test_GetWallet_Handler_ReturnsNotFound(t *testing.T) {
    mockRepo := new(MockWalletRepository)

    mockRepo.On("GetWalletById", "not-existing").Return(nilWallet, app.ErrWalletNotFound)

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

	req := createFakeGetRequest(map[string]string{
		"id": "not-existing",
	})

    rr := httptest.NewRecorder()

    controller.GetWallet(rr, req)

    assert.Equal(t, http.StatusNotFound, rr.Code)
    assert.JSONEq(t, `{"error":"Wallet not found"}`, rr.Body.String())
    mockRepo.AssertExpectations(t)
}

func Test_UpdateWallet_Deposit_Success(t *testing.T) {
    mockRepo := new(MockWalletRepository)

    wallet := &app.Wallet{Id: "1", Amount: 150.0}

    mockRepo.On("Deposit", wallet.Id, 50.0).Return(wallet, nil)

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

	req := createFakePostRequest(map[string]any{
		"walletId": wallet.Id,
		"operationType": "DEPOSIT",
		"amount": 50.0,
	})

    rr := httptest.NewRecorder()

    controller.UpdateWallet(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.JSONEq(t, `{"id":"1","amount":150}`, rr.Body.String())
    mockRepo.AssertExpectations(t)
}

func Test_UpdateWallet_ValidationError(t *testing.T) {
    mockRepo := new(MockWalletRepository)

    mockRepo.On("Deposit", "", 100.0).Return(nilWallet, &app.ValidationError{
        Field:   "id",
        Message: "cannot be empty",
    })

    req := createFakePostRequest(map[string]any{
        "walletId":      "",
        "operationType": "DEPOSIT",
        "amount":        100.0,
    })

    rr := httptest.NewRecorder()

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

    controller.UpdateWallet(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.JSONEq(t, `{"error":"Validation error on field 'id': cannot be empty"}`, rr.Body.String())
    mockRepo.AssertExpectations(t)
}

func Test_UpdateWallet_Withdraw_EmptyID(t *testing.T) {
    mockRepo := new(MockWalletRepository)

    mockRepo.On("Withdraw", "", 50.0).Return(nilWallet, &app.ValidationError{
        Field:   "id",
        Message: "cannot be empty",
    })

    req := createFakePostRequest(map[string]any{
        "walletId":      "",
        "operationType": "WITHDRAW",
        "amount":        50.0,
    })

    rr := httptest.NewRecorder()

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

    controller.UpdateWallet(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.JSONEq(t, `{"error":"Validation error on field 'id': cannot be empty"}`, rr.Body.String())
    mockRepo.AssertExpectations(t)
}

func Test_UpdateWallet_CreateNewWallet(t *testing.T) {
    mockRepo := new(MockWalletRepository)

    wallet := &app.Wallet{Id: "new", Amount: 100.0}

    mockRepo.On("Deposit", "", 100.0).Return(wallet, nil)

	req := createFakePostRequest(map[string]any{
		"walletId": "",
		"operationType": "DEPOSIT",
		"amount": 100.0,
	})

    rr := httptest.NewRecorder()

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

    controller.UpdateWallet(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
    assert.JSONEq(t, `{"id":"new","amount":100}`, rr.Body.String())
    mockRepo.AssertExpectations(t)
}

func Test_UpdateWallet_UnsupportedOperation(t *testing.T) {
    mockRepo := new(MockWalletRepository)

	req := createFakePostRequest(map[string]any{
		"walletId": "1",
        "operationType": "TRANSFER",
        "amount": 100.0,
	})

    rr := httptest.NewRecorder()

    controller := &app.WalletsController{
        WalletsRepository: mockRepo,
    }

    controller.UpdateWallet(rr, req)

    assert.Equal(t, http.StatusBadRequest, rr.Code)
    assert.JSONEq(t, `{"error":"Invalid operation type"}`, rr.Body.String())
}
