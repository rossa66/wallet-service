package api

import (
	"net/http"
	"wallet-service/services"

	"github.com/gorilla/mux"
)

func SetupRoutes(service *services.WalletService) *mux.Router {
	r := mux.NewRouter()
	handler := NewWalletHandler(service)

	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/wallet", handler.HandleOperation).Methods("POST")
	v1.HandleFunc("/wallets/{walletId}", handler.HandleGetBalance).Methods("GET")

	return r
}
