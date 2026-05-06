package exceptions

import "errors"

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrUserDisabled      = errors.New("用户已被禁用")
	ErrEmailNotVerified  = errors.New("邮箱未验证")
	ErrUserAlreadyLogout = errors.New("用户已登出")
	ErrUserNotFoundInDB  = errors.New("USER_NOT_FOUND_IN_DB")
	ErrDuplicPhone       = errors.New("DUPLIC_PHONE")
	ErrRandCodeExpire    = errors.New("RAND_CODE_EXPIRE")
)
