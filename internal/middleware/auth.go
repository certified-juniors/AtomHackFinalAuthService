package middleware

import (
	"github.com/certified-juniors/AtomHack/internal/domain"
	"net/http"
	"time"
)

type AuthMiddleware struct {
	authUsecase domain.AuthUsecase
}

func NewAuth(au domain.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: au}
}

func (m *AuthMiddleware) IsAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				domain.WriteError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			domain.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if c.Expires.After(time.Now()) {
			domain.WriteError(w, "cookie is expired", http.StatusUnauthorized)
		}

		sessionToken := c.Value
		id, err := m.authUsecase.GetUserID(sessionToken)
		if err != nil {
			domain.WriteError(w, err.Error(), domain.GetStatusCode(err))
			return
		}
		if id == "" {
			domain.WriteError(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
