package services

import "context"

type SendPhoneCodeInput struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,min=1,max=12"`
}

type SendEmailVerifyCodeInput struct {
	EmailAddr string `json:"emailAddr" validate:"required,min=1,max=50"`
}

type CommonServices interface {
	SendPhoneCode(ctx context.Context, input SendPhoneCodeInput) int
	SendEmailVerifyCode(ctx context.Context, input SendEmailVerifyCodeInput) int
}
