package model

import "goly-app/database"

func GetAllGolies() ([]Goly, error) {
	var golies []Goly

	tx := database.DB.Find(&golies)
	if tx.Error != nil {
		return []Goly{}, tx.Error
	}

	return golies, nil
}

func GetGoly(id uint64) (Goly, error) {
	var goly Goly

	tx := database.DB.Where("id = ?", id).First(&goly)

	if tx.Error != nil {
		return Goly{}, tx.Error
	}

	return goly, nil
}

func CreateGoly(goly Goly) error {
	tx := database.DB.Create(&goly)
	return tx.Error
}

func UpdateGoly(goly Goly) error {

	tx := database.DB.Save(&goly)
	return tx.Error
}

func DeleteGoly(id uint64) error {

	tx := database.DB.Unscoped().Delete(&Goly{}, id)
	return tx.Error
}

func FindByGolyUrl(url string) (Goly, error) {
	var goly Goly
	tx := database.DB.Where("goly = ?", url).First(&goly)
	return goly, tx.Error
}