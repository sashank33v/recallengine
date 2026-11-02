package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/sashank33v/recallengine/internal/dto"
	"github.com/sashank33v/recallengine/internal/middleware"
	"github.com/sashank33v/recallengine/internal/service"
)

type Authhandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *Authhandler {
	return &Authhandler{authService: authService}
}

func (h *Authhandler) Signup(w http.ResponseWriter, r *http.Request) {
	var request dto.SignupRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, err := h.authService.Signup(r.Context(), request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrEmailTaken):
			writeError(w, http.StatusConflict, err.Error())
		default:
			log.Println("signup error:", err)
			writeError(w, http.StatusInternalServerError, "something went wrong")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
	})
}

func (h *Authhandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		log.Println("login error:", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, dto.LoginResponse{Token: token})
}

func (h *Authhandler) Me(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"user_id": userId})

}
