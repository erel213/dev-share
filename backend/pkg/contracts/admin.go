package contracts

import "github.com/google/uuid"

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
}
