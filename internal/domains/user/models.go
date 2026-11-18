package user

import "time"

type User struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	Password       string     `json:"-"`
	DisplayName    string     `json:"display_name"`
	EmailConfirmed *time.Time `json:"email_confirmed,omitempty"`
	RoleID         int64      `json:"role_id"`
	Role           *Role      `json:"role,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"-"`
}

type userCursor struct {
	CreatedAt time.Time `json:"c"`
	ID        int64     `json:"i"`
}

type Role struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	IsDefault   bool         `json:"is_default"`
	Description *string      `json:"description,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
