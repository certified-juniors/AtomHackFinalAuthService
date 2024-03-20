package usecase

import (
	"bytes"
	"crypto/rand"
	"time"

	"github.com/certified-juniors/AtomHack/internal/domain"
	logs "github.com/certified-juniors/AtomHack/internal/logger"

	"github.com/golang-jwt/jwt"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/argon2"
)

type authUsecase struct {
	authRepo    domain.AuthRepository
	sessionRepo domain.SessionRepository
	jwtSecret   []byte
}

func NewAuthUsecase(ar domain.AuthRepository, sr domain.SessionRepository, js []byte) domain.AuthUsecase {
	return &authUsecase{
		authRepo:    ar,
		sessionRepo: sr,
		jwtSecret:   js,
	}
}

func (u *authUsecase) Login(credentials domain.Credentials) (domain.Session, int, error) {
	expectedUser, err := u.authRepo.GetByEmail(credentials.Email)
	if err != nil {
		return domain.Session{}, 0, err
	}
	logs.Logger.Debug("Usecase Login expected user:", expectedUser)

	if !checkPasswords(expectedUser.Password, credentials.Password) {
		return domain.Session{}, 0, domain.ErrWrongCredentials
	}

	t, err := u.GenerateJWT(credentials.Email)
	if err != nil {
		return domain.Session{}, 0, err
	}

	logs.Logger.Debug("usecase/http Login jwt:\n", t)

	session := domain.Session{
		Token:     t,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserID:    expectedUser.ID,
	}
	if err = u.sessionRepo.Add(session); err != nil {
		return domain.Session{}, 0, err
	}

	return session, expectedUser.ID, nil
}

func (u *authUsecase) Logout(token string) error {
	if token == "" {
		return domain.ErrInvalidToken
	}

	if err := u.sessionRepo.DeleteByToken(token); err != nil {
		return err
	}

	return nil
}

func (u *authUsecase) Register(user domain.User) (int, error) {
	if cmp.Equal(user, domain.User{}) {
		return 0, domain.ErrBadRequest
	}

	if exists, err := u.authRepo.UserExists(user.Email); exists && err == nil {
		return 0, domain.ErrAlreadyExists
	}

	salt := make([]byte, 8)
	rand.Read(salt)
	user.Password = HashPassword(salt, user.Password)
	if id, err := u.authRepo.AddUser(user); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (u *authUsecase) AddCodeByID(id int, code string) error {
	if id <= 0 || code == "" {
		return domain.ErrBadRequest
	}

	err := u.sessionRepo.AddCodeByID(id, code)
	if err != nil {
		return err
	}

	return nil
}

func (u *authUsecase) GetByID(id int) (domain.User, error) {
	if id == 0 {
		return domain.User{}, domain.ErrBadRequest
	}

	user, err := u.authRepo.GetByID(id)
	logs.Logger.Debug("Usecase GetByID user: ", user)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *authUsecase) GetUserID(token string) (string, error) {
	if token == "" {
		return "", domain.ErrInvalidToken
	}

	id, err := u.sessionRepo.GetUserID(token)
	logs.Logger.Debug("Usecase GetUserID auth: ", id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (u *authUsecase) GenerateJWT(email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email

	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//func (u *authUsecase) ParseJWT(tokenString string) (string, error) {
//	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//		return u.jwtSecret, nil
//	})
//
//	if err != nil {
//		return "", err
//	}
//
//	return token.Raw, nil
//}

func HashPassword(salt []byte, password []byte) []byte {
	hashedPass := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

func checkPasswords(passHash []byte, plainPassword []byte) bool {
	salt := make([]byte, 8)
	_ = copy(salt, passHash)
	userPassHash := HashPassword(salt, plainPassword)
	return bytes.Equal(userPassHash, passHash)
}
