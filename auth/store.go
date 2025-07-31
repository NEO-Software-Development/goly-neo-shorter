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

func DeleteSessionByToken(token string) error {
	tx := database.DB.Where("token = ?", token).Delete(&Session{})
	return tx.Error
}

func CreateSession(session *Session) error {
	tx := database.DB.Create(session)
	return tx.Error
}

func GetSessionByToken(token string) (*Session, error) {
	var session Session
	tx := database.DB.Where("token = ?", token).First(&session)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &session, nil
}

func GetUserByID(id uint) (*User, error) {
	var user User
	tx := database.DB.Where("id = ?", id).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}
