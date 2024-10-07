package models

import (
	"errors"
	"fmt"

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

var (
	ErrEmailInUse = errors.New("email already in use")
)

func NewUserModel(db *gorm.DB) *UserModel {
	return &UserModel{DB: db}
}

func (um *UserModel) InsertUser(email, username, password string) (*User, error) {
	var user User
	// Check if the email is already in use
	res := um.DB.Where("email = ?", email).Limit(1).Find(&user)
	if res.RowsAffected > 0 {
		return nil, ErrEmailInUse
	}

	hp, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user = User{
		Username: username,
		Password: string(hp),
		Email:    email,
	}
	res = um.DB.Create(&user)
	if res.Error != nil {
		fmt.Println(res.Error.Error())
		return nil, res.Error
	}

	return &user, nil
}

func (um *UserModel) LoginUser(email, password string) (*User, error) {
	var user User
	res := um.DB.Where("email = ?", email).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
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
