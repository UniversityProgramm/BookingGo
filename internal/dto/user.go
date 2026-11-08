package dto

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
	FIO      string `json:"fio" binding:"required"`
	Phone    string `json:"phone" binding:"required,min=10,max=15"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=20"`
}

type UpdateUserProfileRequest struct {
	FIO   *string `json:"fio"`
	Phone *string `json:"phone,min=10,max=15"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"cur_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=20"`
	OtpCode         string `json:"otp_code" binding:"required,len=6"`
}

type ChangeEmailRequest struct {
	NewEmail        string `json:"new_email" binding:"required,email"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	OtpCode         string `json:"otp_code" binding:"required,len=6"`
}

type VerifyTotpRequest struct {
	OtpCode string `json:"otp_code" binding:"required,len=6"`
}

type DisableTotpRequest struct {
	OtpCode         string `json:"otp_code" binding:"required,len=6"`
	CurrentPassword string `json:"confirm_password" binding:"required"`
}
