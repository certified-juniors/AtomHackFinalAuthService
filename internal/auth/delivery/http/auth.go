package http

import (
	"crypto/rand"
	"encoding/json"
	"github.com/certified-juniors/AtomHack/internal/auth/delivery/smtp"
	"math/big"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/certified-juniors/AtomHack/internal/domain"
	logs "github.com/certified-juniors/AtomHack/internal/logger"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	AuthUsecase domain.AuthUsecase
}

func NewAuthHandler(authMwRouter *mux.Router, mainRouter *mux.Router, u domain.AuthUsecase) {
	handler := &AuthHandler{
		AuthUsecase: u,
	}

	mainRouter.HandleFunc("/api/v1/auth/login", handler.Login).Methods(http.MethodPost, http.MethodOptions)
	mainRouter.HandleFunc("/api/v1/auth/register", handler.Register).Methods(http.MethodPost, http.MethodOptions)
	mainRouter.HandleFunc("/api/v1/auth/confirm", handler.Confirm).Methods(http.MethodPost, http.MethodOptions)
	mainRouter.HandleFunc("/api/v1/auth/me", handler.Me).Methods(http.MethodGet, http.MethodOptions)

	authMwRouter.HandleFunc("/v1/auth/check", handler.CheckAuth).Methods(http.MethodPost, http.MethodOptions)
	authMwRouter.HandleFunc("/v1/auth/logout", handler.Logout).Methods(http.MethodPost, http.MethodOptions)
}

// Login godoc
//
//	@Summary		login user
//	@Description	create user session and put it into cookie
//	@Tags			Auth
//	@Accept			json
//	@Param			body	body		domain.Credentials	true	"user credentials"
//	@Success		200		{object}	object{body=object{id=int}}
//	@Failure		400		{object}	object{err=string}
//	@Failure		403		{object}	object{err=string}
//	@Failure		404		{object}	object{err=string}
//	@Failure		500		{object}	object{err=string}
//	@Router			/api/v1/auth/login [post]
func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	auth, err := a.getUserID(r)
	if auth != 0 {
		domain.WriteError(w, "you must be unauthorised", domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Login", err, "User is already logged in")
		return
	}

	var credentials domain.Credentials

	err = json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		domain.WriteError(w, err.Error(), http.StatusBadRequest)
		logs.LogError(logs.Logger, "auth/http", "Login", err, "Failed to decode json from body")
		return
	}
	logs.Logger.Debug("Login credentials:", credentials)
	defer domain.CloseAndAlert(r.Body, "auth/http", "Login")

	if err = checkCredentials(credentials); err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Login", err, "Credentials are incorrect")
		return
	}
	credentials.Email = strings.TrimSpace(credentials.Email)

	session, userID, err := a.AuthUsecase.Login(credentials)
	if err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Login", err, "Failed to login")
		return
	}
	logs.Logger.Debug("auth/http Login: session:", session)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	domain.WriteResponse(
		w,
		map[string]interface{}{
			"id": userID,
		},
		http.StatusOK,
	)
}

// Logout godoc
//
//	@Summary		logout user
//	@Description	delete current session and nullify cookie
//	@Tags			Auth
//	@Success		204
//	@Failure		400	{object}	object{err=string}
//	@Failure		403	{object}	object{err=string}
//	@Failure		404	{object}	object{err=string}
//	@Failure		500	{object}	object{err=string}
//	@Router			/api/v1/auth/logout [post]
func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	sessionToken := c.Value
	logs.Logger.Debug("auth/http Logout: session token:", c)

	if err = a.AuthUsecase.Logout(sessionToken); err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Logout", err, "Failed to logout")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	w.WriteHeader(http.StatusNoContent)
}

// Register godoc
//
//	@Summary		register user
//	@Description	add new user to db and return it id
//	@Tags			Auth
//	@Produce		json
//	@Accept			json
//	@Param			body	body		domain.UserWithoutId	true	"user credentials"
//	@Success		200		{object}	object{body=object{id=int}}
//	@Failure		400		{object}	object{err=string}
//	@Failure		403		{object}	object{err=string}
//	@Failure		500		{object}	object{err=string}
//	@Router			/api/v1/auth/register [post]
func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	auth, err := a.getUserID(r)
	if auth != 0 {
		domain.WriteError(w, "you must be unauthorised", domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Register.getUserID", err, "user is authorised")
		return
	}

	var user domain.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		domain.WriteError(w, "you must be unauthorised", http.StatusBadRequest)
		logs.LogError(logs.Logger, "auth/http", "Register.decode", err, "Failed to decode json from body")
		return
	}
	//logs.Logger.Debug("Register user:", user)
	defer domain.CloseAndAlert(r.Body, "auth/http", "Register")

	user.Email = strings.TrimSpace(user.Email)
	if err = checkCredentials(domain.Credentials{Email: user.Email, Password: user.Password}); err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Register.credentials", err, "creds are invalid")
		return
	}

	var id int
	if id, err = a.AuthUsecase.Register(user); err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Register.register", err, "Failed to register")
		return
	}

	code, err := generateRandomNumber()
	if err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Register.register", err, "Failed to generate code")
		return
	}

	err = smtp.SendMailToClient("Код подтверждения", code, user.Email)
	if err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Register.register", err, "Failed to send code")
		return
	}

	err = a.AuthUsecase.AddCodeByID(id, code)
	if err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Register.register", err, "Failed to save code")
		return
	}

	domain.WriteResponse(
		w,
		map[string]interface{}{
			"id": id,
		},
		http.StatusOK,
	)
}

// Confirm godoc
//
//	@Summary		confirm user
//	@Description	confirm user
//	@Param			body	body		domain.ConfirmPair	true	"user id and verification code"
//	@Tags			Auth
//	@Success		204
//	@Failure		400	{object}	object{err=string}
//	@Failure		500	{object}	object{err=string}
//	@Router			/api/v1/auth/confirm [post]
func (a *AuthHandler) Confirm(w http.ResponseWriter, r *http.Request) {

	var cp domain.ConfirmPair
	err := json.NewDecoder(r.Body).Decode(&cp)
	if err != nil {
		domain.WriteError(w, "somethings wrong with JSON", http.StatusBadRequest)
		logs.LogError(logs.Logger, "auth/http", "CheckAuth", err, "Failed to decode json from body")
		return
	}

	var session domain.Session
	if session, err = a.AuthUsecase.ConfirmUser(cp); err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "CheckAuth", err, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	w.WriteHeader(http.StatusNoContent)
}

// CheckAuth godoc
//
//	@Summary		check getUserID
//	@Description	check if user is authenticated
//	@Tags			Auth
//	@Success		204
//	@Failure		400	{object}	object{err=string}
//	@Failure		401	{object}	object{err=string}
//	@Failure		409	{object}	object{err=string}
//	@Failure		500	{object}	object{err=string}
//	@Router			/api/v1/auth/check [post]
func (a *AuthHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	auth, err := a.getUserID(r)
	if auth != 0 {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "http", "CheckAuth", err, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me godoc
//
//	@Summary		returns user data
//	@Description	returns user data
//	@Tags			Auth
//	@Success		200		{object}	object{body=object{user=domain.UserWithoutPassword}}
//	@Failure		400	{object}	object{err=string}
//	@Failure		401	{object}	object{err=string}
//	@Failure		409	{object}	object{err=string}
//	@Failure		500	{object}	object{err=string}
//	@Router			/api/v1/auth/me [get]
func (a *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	id, err := a.getUserID(r)
	if id == 0 {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Me", err, err.Error())
		return
	}

	var user domain.User
	if user, err = a.AuthUsecase.GetByID(id); err != nil {
		domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
		logs.LogError(logs.Logger, "auth/http", "Me", err, err.Error())
		return
	}

	domain.WriteResponse(
		w,
		map[string]interface{}{
			"user": user,
		},
		http.StatusOK,
	)
}

func generateRandomNumber() (string, error) {
	const digits = "0123456789"
	const length = 6

	var randomNumber string
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}

		randomDigit := digits[randomIndex.Int64()]

		randomNumber += string(randomDigit)
	}

	return randomNumber, nil
}

func (a *AuthHandler) getUserID(r *http.Request) (int, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return 0, domain.ErrUnauthorized
		}

		return 0, domain.ErrBadRequest
	}
	if c.Expires.After(time.Now()) {
		return 0, domain.ErrUnauthorized
	}
	sessionToken := c.Value
	strID, err := a.AuthUsecase.GetUserID(sessionToken)
	if err != nil {
		return 0, err
	}
	if strID == "" {
		return 0, domain.ErrUnauthorized
	}

	id, err := strconv.Atoi(strID)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}
	return id, domain.ErrAlreadyExists
}

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func checkCredentials(cred domain.Credentials) error {
	if cred.Email == "" || len(cred.Password) == 0 {
		return domain.ErrWrongCredentials
	}

	if !valid(cred.Email) {
		return domain.ErrWrongCredentials
	}

	return nil
}
