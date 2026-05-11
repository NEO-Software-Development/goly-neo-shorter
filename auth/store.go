package auth

import (
	"goly-app/database"
	"time"

	"gorm.io/gorm"
)

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

// PurgeUser hard-deletes the user, all their sessions, backup codes, and
// verification tokens in one transaction. Directory data is purged separately
// via directory.PurgeOwner before this is called.
func PurgeUser(userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&Session{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&BackupCode{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&VerificationToken{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Where("id = ?", userID).Delete(&User{}).Error
	})
}

func UpdateUser(user *User) error {
	return database.DB.Save(user).Error
}

func ListSessionsForUser(userID uint) ([]Session, error) {
	var out []Session
	tx := database.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now()).Order("created_at DESC").Find(&out)
	return out, tx.Error
}

func DeleteSessionForUser(id, userID uint) error {
	return database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&Session{}).Error
}
