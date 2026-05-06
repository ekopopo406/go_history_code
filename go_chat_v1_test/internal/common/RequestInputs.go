package common

// CreateUserInput 创建用户的输入参数
type CreateUserInput struct {
	Name        string `json:"name" validate:"min=1,max=50"`
	PhoneNumber string `json:"phoneNumber" validate:"required,min=1,max=50"`
	Email       string `json:"email" validate:"required,email,max=100"`
	Password    string `json:"password" validate:"required,min=1,max=50"`
	RandCode    string `json:"randCode" validate:"required,min=1,max=4"`
}

type CreateUserInputWithPhone struct {
	PhoneNumber string
}

type UpdateUserInput struct {
	ID   uint
	Name string
	Age  int
}

type SearchUserInput struct {
	Page     int    `json:"page" validate:"required,min=1,max=50"`
	PageSize int    `json:"pageSize" validate:"required,min=1,max=50"`
	Name     string `json:"name"`
}
