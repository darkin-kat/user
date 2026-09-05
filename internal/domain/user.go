package domain

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
	RoleWarehouse Role = "warehouse"
)

type User struct {
	ID           string    `bson:"_id,omitempty"`
	FirstName    string    `bson:"first_name"`
	LastName     string    `bson:"last_name"`
	Email        string    `bson:"email"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
	Role         Role      `bson:"role"`
}
