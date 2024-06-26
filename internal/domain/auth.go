package domain

import (
	"time"
)

type Role int

const (
	Usr Role = iota
	Moder
)

type Key string

const SessionContextKey Key = "SessionContextKey"

type SessionContext struct {
	UserID int
	Role   Role
}

type Credentials struct {
	Password []byte `json:"password"`
	Email    string `json:"email"`
}

type User struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	Password   []byte `json:"password,omitempty"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	MiddleName string `json:"middleName"`
	Role       string `json:"role"`
	Confirmed  bool   `json:"confirmed"`
}

type Session struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UserID    int       `json:"-"`
}

type SMTPParams struct {
	SmtpHost        string
	SmtpPort        int
	NoreplyUsername string
	NoreplyPassword string
	DestUsername    string
	DestPassword    string
}

type ConfirmPair struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
}

type AuthUsecase interface {
	Login(credentials Credentials) (Session, int, error)
	Logout(token string) error
	Register(user User) (int, error)
	GetUserID(token string) (string, error)
	GenerateJWT(email string) (string, error)
	GetByID(id int) (User, error)
	AddCodeByID(id int, code string) error
	ConfirmUser(pair ConfirmPair) (Session, error)
}

type AuthRepository interface {
	GetByEmail(email string) (User, error)
	GetByID(id int) (User, error)
	AddUser(user User) (int, error)
	UserExists(email string) (bool, error)
	ConfirmUser(id int) (string, error)
}

type SessionRepository interface {
	Add(session Session) error
	DeleteByToken(token string) error
	GetUserID(token string) (string, error)
	AddCodeByID(id int, code string) error
	GetCodeByID(id string) (string, error)
}
