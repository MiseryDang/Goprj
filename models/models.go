package models

import "github.com/dgrijalva/jwt-go"

type Movie struct {
	ID         string    `json:"id"`
	Isbn       string    `json:"isbn"`
	Title      string    `json:"title"`
	Director   *Director `json:"director"`
	DirectorID uint      `json:"director_id" gorm:"foreignKey:DirectorID"`
	CreatedBy  uint      `json:"created_by"`
}

type Director struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type Registration struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique"`
	FullName string `json:"fullname"`
	Password string `json:"password"`
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

type TokenStore struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Token string `json:"token" gorm:"unique"`
}
