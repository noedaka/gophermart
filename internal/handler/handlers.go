package handler

import (
	"encoding/json"
	"errors"
	"gophermart/internal/config"
	"gophermart/internal/model"
	"gophermart/internal/service"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var credentials *model.UserCredentials

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.service.Register(r.Context(), *credentials)
	if err != nil {
		http.Error(w, "Registration error", http.StatusInternalServerError)
		return
	}

	if userID == -1 {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials *model.UserCredentials

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.service.Login(r.Context(), *credentials)
	if err != nil {
		http.Error(w, "Login error", http.StatusInternalServerError)
		return
	}

	if userID == -1 {
		http.Error(w, "Wrong login or password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(config.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(bodyBytes) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	code := string(bodyBytes)

	if !service.Luhn(code) {
		http.Error(w, "Error reading request body", http.StatusUnprocessableEntity)
		return
	}

	err = h.service.CreateOrder(r.Context(), userID, code)

	switch err {
	case nil:
		w.WriteHeader(http.StatusAccepted)
		return
	case config.ErrOrderAlreadyUploadedByUser:
		w.WriteHeader(http.StatusOK)
		return
	case config.ErrOrderAlreadyUploadedByAnother:
		http.Error(w, err.Error(), http.StatusConflict)
		return
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(config.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetOrders(r.Context(), userID)
	if errors.Is(err, config.ErrNoOrders) {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(config.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetBalance(r.Context(), userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateWithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(config.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var withdrawal model.Withdrawal

	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !service.Luhn(withdrawal.Number) {
		http.Error(w, "not correct code", http.StatusUnprocessableEntity)
		return
	}

	err := h.service.CreateWithdrawal(r.Context(), userID, withdrawal)

	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
		return
	case config.ErrNotEnoughMoney:
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	case config.ErrOrderAlreadyUploadedByUser:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(config.UserIDKey).(int64)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.service.GetWithdrawals(r.Context(), userID)
	if errors.Is(err, config.ErrNoOrders) {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
