package errs

type CodeError struct {
	Code int
	Msg  string
}

func (e *CodeError) Error() string {
	return e.Msg
}

var (
	ErrorTodoNotFound = &CodeError{
		Code: 404,
		Msg:  "备忘录不存在",
	}
	ErrorTodoAlreadyExists = &CodeError{
		Code: 409,
		Msg:  "备忘录已存在",
	}
	ErrorTodoNull = &CodeError{
		Code: 400,
		Msg:  "备忘录名不能为空",
	}
	ErrorUserNotFound = &CodeError{
		Code: 404,
		Msg:  "用户不存在",
	}
	ErrorUserAlreadyExists = &CodeError{
		Code: 409,
		Msg:  "用户名/邮箱已被注册",
	}
	ErrorEmailAlreadyExists = &CodeError{
		Code: 409,
		Msg:  "邮箱已被注册",
	}
	ErrorService = &CodeError{
		Code: 500,
		Msg:  "服务器内部错误",
	}
	ErrorUserNull = &CodeError{
		Code: 400,
		Msg:  "用户名/密码/邮箱不能为空",
	}
	ErrorInvalidToken = &CodeError{
		Code: 401,
		Msg:  "登录已失效，请重新登录",
	}
	ErrorInvalidCredentials = &CodeError{
		Code: 401,
		Msg:  "用户名或密码错误"}
)
