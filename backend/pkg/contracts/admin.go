package contracts

import (
	"time"

	"github.com/google/uuid"
)

type AdminInit struct {
	AdminName            string `json:"admin_name" validate:"required,min=2,max=100"`
	AdminEmail           string `json:"admin_email" validate:"required,email"`
	AdminPassword        string `json:"admin_password" validate:"required,min=8,strongpassword"`
	WorkspaceName        string `json:"workspace_name" validate:"required,min=3,max=100"`
	WorkspaceDescription string `json:"workspace_description" validate:"max=500"`
}

type AdminInitResponse struct {
	Message     string    `json:"message"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	AdminUserID uuid.UUID `json:"admin_user_id"`
	UserName    string    `json:"admin_user_name"`
}

type InviteUser struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin editor user"`
}

type InviteUserResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Password string    `json:"password"`
}

type ResetPassword struct {
	UserID uuid.UUID `json:"user_id" validate:"required,uuid4"`
}

type ResetPasswordResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password"`
}

type AdminUserResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
