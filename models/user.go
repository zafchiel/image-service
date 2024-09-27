package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"not null"`
	Password string `gorm:"not null"`
	Email    string `gorm:"unique;uniqueIndex;not null"`
	Images   []ImageMetadata
}

type UserModel struct {
	DB *gorm.DB
}

func (um *UserModel) InsertUser(email, username, password string) error {
	hp, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	res := um.DB.Create(&User{
		Username: username,
		Password: string(hp),
		Email:    email,
	})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (um *UserModel) LoginUser(email, hashedPassword string) (*User, error) {
	var user User
	res := um.DB.Where("email = ?", email).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (um *UserModel) GetUserByEmail(email string) (*User, error) {
	var user User

	res := um.DB.Where("email = ?", email).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}
