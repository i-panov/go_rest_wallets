package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"go_rest_wallets/app"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {
		log.Fatal("Не установлена переменная окружения DATABASE_URL")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)

	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	defer pool.Close()

	fmt.Println("Успешное подключение к PostgreSQL!")

	walletsRepository := app.WalletsRepository{Conn: pool}
	walletsController := app.WalletsController{WalletsRepository: &walletsRepository}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/wallets/{id}", walletsController.GetWallet).Methods("GET")
	r.HandleFunc("/api/v1/wallet", walletsController.UpdateWallet).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}
