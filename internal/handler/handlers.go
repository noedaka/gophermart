package handler

import (
	"encoding/json"
	"gophermart/internal/config"
	"gophermart/internal/model"
	"gophermart/internal/service"
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
