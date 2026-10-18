// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"todo/internal/logic"
	"todo/internal/svc"
	"todo/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUserLoginLogic(r.Context(), svcCtx)
		resp, err := l.UserLogin(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    resp.Token,
				Path:     "/",
				MaxAge:   svcCtx.Config.Token.ExpireSeconds,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   svcCtx.Config.Token.CookieSecure,
			})
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
