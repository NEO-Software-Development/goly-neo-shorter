package model


// GetAllGolies retrieves all Goly links from the database.
// It returns a slice of Goly structs and an error if the query fails.
func GetAllGolies() ([]Goly, error) {
	var golies []Goly

	tx := db.Find(&golies)
	if tx.Error != nil {
		return []Goly{}, tx.Error
	}

	return golies, nil
}

// GetGoly retrieves a single Goly link from the database by its ID.
// It takes a uint64 ID as a parameter.
// It returns a Goly struct and an error if the query fails.
func GetGoly(id uint64) (Goly, error) {
	var goly Goly

	tx := db.Where("id = ?", id).First(&goly)

	if tx.Error != nil {
		return Goly{}, tx.Error
	}

	return goly, nil
}

// CreateGoly creates a new Goly link in the database.
// It takes a Goly struct as a parameter.
// It returns an error if the creation fails.
func CreateGoly(goly Goly) error {
	tx := db.Create(&goly)
	return tx.Error
}

// UpdateGoly updates an existing Goly link in the database.
// It takes a Goly struct as a parameter.
// It returns an error if the update fails.
func UpdateGoly(goly Goly) error {

	tx := db.Save(&goly)
	return tx.Error
}

// DeleteGoly deletes a Goly link from the database by its ID.
// It takes a uint64 ID as a parameter.
// It returns an error if the deletion fails.
func DeleteGoly(id uint64) error {

	tx := db.Unscoped().Delete(&Goly{}, id)
	return tx.Error
}

// FindByGolyUrl retrieves a single Goly link from the database by its "goly" URL.
// It takes a string URL as a parameter.
// It returns a Goly struct and an error if the query fails.
func FindByGolyUrl(url string) (Goly, error) {
	var goly Goly
	tx := db.Where("goly = ?", url).First(&goly)
	return goly, tx.Error
}