package handler

import (
	"errors"
	"net/http"
	"todo/internal/errs"
)

func HandleError(err error) (int, any) {
	var codeErr *errs.CodeError
	if errors.As(err, &codeErr) {
		return codeErr.Code, codeErr.Msg
	}
	return http.StatusBadRequest, err.Error()
}
