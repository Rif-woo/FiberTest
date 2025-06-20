package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username  string    `gorm:"type:varchar(50);unique;not null" validate:"required,min=3,max=50"`
	Email     string    `gorm:"type:varchar(100);unique;not null" validate:"required,email"`
	Password  string    `gorm:"type:text;not null" validate:"required,min=8"`
	Role      string    `gorm:"type:user_role;default:'user';not null" validate:"required,oneof=admin user"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}


//Oauth with fiber golang
