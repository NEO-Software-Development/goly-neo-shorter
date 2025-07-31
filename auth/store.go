package auth

import "goly-app/database"

func CreateUser(user *User) error {
	tx := database.DB.Create(user)
	return tx.Error
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	tx := database.DB.Where("username = ?", username).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}
