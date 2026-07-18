package auth

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

type RegisterRequest struct {
	DisplayName          string `json:"display_name"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type ConfirmResendRequest struct {
	Email string `json:"email"`
}

type RequestResetRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token                string `json:"token"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type MeResponse struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}
