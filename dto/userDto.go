package dto

// Register DTO
type RegisterUserDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	RoleID   uint   `json:"role_id"` // optional, default role user
}

// Login DTO
type LoginDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileDto struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type ProfileDto struct {
	UserID   uint   `json:"user_id" validate:"required"`
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Address  string `json:"address" validate:"required"`
	User     struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}
