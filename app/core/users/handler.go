package users

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"rest_api_go/auth"
	"rest_api_go/middleware"
)

type UserHandler struct {
	Service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{Service: service}
}


type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	IsActive bool   `json:"is_active"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	user, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessSID := uuid.NewString()
	refreshSID := uuid.NewString()

	if err := auth.SaveSession(accessSID, user.ID, auth.AccessTokenTTL); err != nil {
		log.Printf("redis save access session: %v", err)
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}
	if err := auth.SaveSession(refreshSID, user.ID, auth.RefreshTokenTTL); err != nil {
		log.Printf("redis save refresh session: %v", err)
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, accessSID)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}
	refreshToken, err := auth.GenerateRefreshToken(user.ID, refreshSID)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Refresh issues a new access token when the refresh token is still valid in Redis.
func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	claims, err := auth.ParseToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	if claims.Type != "refresh" {
		http.Error(w, "not a refresh token", http.StatusUnauthorized)
		return
	}
	if claims.ID == "" {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}
	if err := auth.ValidateSession(claims.ID, claims.UserID); err != nil {
		http.Error(w, "session expired or logged out", http.StatusUnauthorized)
		return
	}

	newAccessSID := uuid.NewString()
	if err := auth.SaveSession(newAccessSID, claims.UserID, auth.AccessTokenTTL); err != nil {
		log.Printf("redis save access session: %v", err)
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}

	accessToken, err := auth.GenerateAccessToken(claims.UserID, newAccessSID)
	if err != nil {
		http.Error(w, "could not issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}

// Logout removes the current access session from Redis (client should discard tokens).
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sid, ok := middleware.SessionIDFromContext(r.Context())
	if !ok || sid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := auth.DeleteSession(sid); err != nil {
		log.Printf("redis delete session: %v", err)
		http.Error(w, "could not log out", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// Handler methods (Create, Get, Update, Delete, List) would go here
// func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

// 	// 1. Check content type
// 	if r.Header.Get("Content-Type") != "application/json" {
// 		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
// 		return
// 	}

// 	var user User

// 	// 2. Decode request
// 	err := json.NewDecoder(r.Body).Decode(&user)
// 	if err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	// 3. Set timestamps (good)
// 	user.CreatedAt = time.Now()
// 	user.UpdatedAt = time.Now()

// 	// 4. Call service
// 	err = h.Service.CreateUser(&user)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// 5. Response
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(user)
// }

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var req CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	user := User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		IsActive: req.IsActive,
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err = h.Service.CreateUser(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
// 	// Implementation for getting a user by ID
// 	// 1. Check content type
// 	if r.Header.Get("Content-Type") != "application/json" {
// 		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
// 		return
// 	}
// 	vars := mux.Vars(r)

// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 		return
// 	}

// 	user, err := h.Service.GetUserByID(id)
// 	if err != nil {
// 		http.Error(w, "User not found", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(user)
// }


func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetUserByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}


// func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
// 	// Implementation for updating a user
// 	// 1. Check content type
// 	if r.Header.Get("Content-Type") != "application/json" {
// 		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
// 		return
// 	}
	
// 	var user User
// 	err := json.NewDecoder(r.Body).Decode(&user)
// 	if err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	user.UpdatedAt = time.Now()

// 	err = h.Service.UpdateUser(&user)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(user)
// }

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user.UpdatedAt = time.Now()

	err = h.Service.UpdateUser(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
// 	// Implementation for deleting a user
// 	vars := mux.Vars(r)
	
// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 		return
// 	}

// 	err = h.Service.DeleteUser(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)

// 	w.WriteHeader(http.StatusNoContent)
// }
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.Service.DeleteUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
// 	// Implementation for listing all users
// 	users, err := h.Service.ListUsers()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(users)
// }


func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {

	users, err := h.Service.ListUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
// func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
// 	// Implementation for deactivating a user (soft delete)
// 	vars := mux.Vars(r)
// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 		return
// 	}
	
// 	err = h.Service.DeactivateUser(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// }

func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.Service.DeactivateUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}



// func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
// 	// Implementation for activating a user
// 	vars := mux.Vars(r)
// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 		return
// 	}
	
// 	err = h.Service.ActivateUser(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// }



func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.Service.ActivateUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}