package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type WalletsController struct {
	WalletsRepository WalletsRepositoryInterface
}

func (c *WalletsController) GetWallet(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars["id"]
	wallet, err := c.WalletsRepository.GetWalletById(id)
	writeResponse(writer, wallet, err)
}

func (c *WalletsController) UpdateWallet(writer http.ResponseWriter, request *http.Request) {
	var requestData struct {
		WalletID      string  `json:"walletId"`
		OperationType string  `json:"operationType"`
		Amount        float64 `json:"amount"`
	}

	if err := json.NewDecoder(request.Body).Decode(&requestData); err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if strings.EqualFold(requestData.OperationType, "DEPOSIT") {
		wallet, err := c.WalletsRepository.Deposit(requestData.WalletID, requestData.Amount)
		writeResponse(writer, wallet, err)
	} else if strings.EqualFold(requestData.OperationType, "WITHDRAW") {
		wallet, err := c.WalletsRepository.Withdraw(requestData.WalletID, requestData.Amount)
		writeResponse(writer, wallet, err)
	} else {
		writeErrorResponse(writer, http.StatusBadRequest, "Invalid operation type")
		return
	}
}

func writeResponse(writer http.ResponseWriter, wallet *Wallet, err error) {
	if err == ErrWalletNotFound {
		writeErrorResponse(writer, http.StatusNotFound, "Wallet not found")
		return
	}

	if err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			writeErrorResponse(writer, http.StatusBadRequest, validationErr.Error())
			return
		}

		writeErrorResponse(writer, http.StatusInternalServerError, err.Error())
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(wallet.ToJson())
}

func writeErrorResponse(writer http.ResponseWriter, statusCode int, message string) {
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write([]byte(`{"error": "` + message + `"}`))
}
