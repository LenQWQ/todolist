package middleware

import (
	"context"
	"net/http"
	"time"
	"todo/internal/errs"
	"todo/internal/svc"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthInfo struct {
	UserID int
	Token  string
}
type ctxKey struct{}

var authKey ctxKey

func AuthInfoFrom(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authKey).(*AuthInfo)
	return info, ok
}

var publicPaths = map[string]bool{
	"/user/register": true,
	"/user/login":    true,
}

func Auth(svcCtx *svc.ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if publicPaths[r.URL.Path] {
				next(w, r)
				return
			}
			cookie, err := r.Cookie("token")
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errs.ErrorInvalidToken)
				return
			}
			tokenStr := cookie.Value
			tk, err := svcCtx.TokenRepo.FindByToken(tokenStr)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errs.ErrorInvalidToken)
				return
			}
			if tk.ExpiresAt < time.Now().Unix() {
				httpx.ErrorCtx(r.Context(), w, errs.ErrorInvalidToken)
				return
			}
			ctx := context.WithValue(r.Context(), authKey, &AuthInfo{UserID: tk.UserId, Token: tokenStr})
			next(w, r.WithContext(ctx))
		}
	}
}
