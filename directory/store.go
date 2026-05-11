package directory

import (
	"errors"
	"goly-app/database"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a directory or link cannot be located within
// the caller's scope. Handlers map it to 404 so we never leak existence.
var ErrNotFound = errors.New("not found")

func CreateDirectory(d *Directory) error {
	return database.DB.Create(d).Error
}

func UpdateDirectory(d *Directory) error {
	return database.DB.Save(d).Error
}

func SoftDeleteDirectory(id, ownerID uint) error {
	tx := database.DB.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&Directory{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// HardDeleteDirectory removes the directory, its links, and audit entries in
// one transaction. Used for GDPR-style erasure.
func HardDeleteDirectory(id, ownerID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Unscoped().Where("id = ? AND owner_id = ?", id, ownerID).Delete(&Directory{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		if err := tx.Unscoped().Where("directory_id = ?", id).Delete(&ContactLink{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Where("directory_id = ?", id).Delete(&AuditEntry{}).Error
	})
}

func ListDirectoriesByOwner(ownerID uint) ([]Directory, error) {
	var out []Directory
	tx := database.DB.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out)
	return out, tx.Error
}

func GetDirectoryForOwner(id, ownerID uint) (*Directory, error) {
	var d Directory
	tx := database.DB.Preload("Links").Where("id = ? AND owner_id = ?", id, ownerID).First(&d)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, tx.Error
}

func GetPublishedDirectoryBySlug(slug string) (*Directory, error) {
	var d Directory
	tx := database.DB.Preload("Links", func(db *gorm.DB) *gorm.DB {
		return db.Where("visibility = ?", "public").Order("position ASC")
	}).Where("slug = ? AND is_published = ?", slug, true).First(&d)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, tx.Error
}

func SlugExists(slug string) (bool, error) {
	var count int64
	tx := database.DB.Model(&Directory{}).Where("slug = ?", slug).Count(&count)
	return count > 0, tx.Error
}

func IncrementViews(slug string) error {
	return database.DB.Model(&Directory{}).
		Where("slug = ? AND is_published = ?", slug, true).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

func CreateLink(l *ContactLink) error {
	return database.DB.Create(l).Error
}

func UpdateLink(l *ContactLink) error {
	return database.DB.Save(l).Error
}

func GetLinkForOwner(linkID, directoryID, ownerID uint) (*ContactLink, error) {
	var l ContactLink
	tx := database.DB.
		Joins("JOIN directories ON directories.id = contact_links.directory_id").
		Where("contact_links.id = ? AND contact_links.directory_id = ? AND directories.owner_id = ?",
			linkID, directoryID, ownerID).
		First(&l)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &l, tx.Error
}

// GetPublicLinkValue fetches a single link's value for visitor reveal-on-tap.
// It enforces published-directory + public-visibility at the query layer; any
// mismatch returns ErrNotFound so the response can be a flat 404.
func GetPublicLinkValue(slug string, linkID uint) (*ContactLink, error) {
	var l ContactLink
	tx := database.DB.
		Joins("JOIN directories ON directories.id = contact_links.directory_id").
		Where("contact_links.id = ? AND contact_links.visibility = ? AND directories.slug = ? AND directories.is_published = ?",
			linkID, "public", slug, true).
		First(&l)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &l, tx.Error
}

func DeleteLink(linkID, directoryID, ownerID uint) error {
	tx := database.DB.
		Where("id = ? AND directory_id = ? AND directory_id IN (SELECT id FROM directories WHERE owner_id = ?)",
			linkID, directoryID, ownerID).
		Delete(&ContactLink{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func ListLinksForOwner(directoryID, ownerID uint) ([]ContactLink, error) {
	var out []ContactLink
	tx := database.DB.
		Joins("JOIN directories ON directories.id = contact_links.directory_id").
		Where("contact_links.directory_id = ? AND directories.owner_id = ?", directoryID, ownerID).
		Order("contact_links.position ASC").
		Find(&out)
	return out, tx.Error
}

func ReorderLinks(directoryID, ownerID uint, positions map[uint]int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var owned int64
		if err := tx.Model(&Directory{}).Where("id = ? AND owner_id = ?", directoryID, ownerID).Count(&owned).Error; err != nil {
			return err
		}
		if owned == 0 {
			return ErrNotFound
		}
		for id, pos := range positions {
			if err := tx.Model(&ContactLink{}).
				Where("id = ? AND directory_id = ?", id, directoryID).
				Update("position", pos).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func WriteAudit(userID, directoryID uint, action, summary string) {
	_ = database.DB.Create(&AuditEntry{
		UserID:      userID,
		DirectoryID: directoryID,
		Action:      action,
		Summary:     summary,
	}).Error
}

func ListAudit(directoryID, ownerID uint, limit int) ([]AuditEntry, error) {
	var ok int64
	if err := database.DB.Model(&Directory{}).Where("id = ? AND owner_id = ?", directoryID, ownerID).Count(&ok).Error; err != nil {
		return nil, err
	}
	if ok == 0 {
		return nil, ErrNotFound
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var out []AuditEntry
	tx := database.DB.Where("directory_id = ?", directoryID).Order("created_at DESC").Limit(limit).Find(&out)
	return out, tx.Error
}

// PurgeOwner hard-deletes everything tied to an owner. Used by DELETE /api/v1/me.
func PurgeOwner(ownerID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var dirIDs []uint
		if err := tx.Model(&Directory{}).Unscoped().Where("owner_id = ?", ownerID).Pluck("id", &dirIDs).Error; err != nil {
			return err
		}
		if len(dirIDs) > 0 {
			if err := tx.Unscoped().Where("directory_id IN ?", dirIDs).Delete(&ContactLink{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("directory_id IN ?", dirIDs).Delete(&AuditEntry{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("owner_id = ?", ownerID).Delete(&Directory{}).Error; err != nil {
				return err
			}
		}
		return tx.Unscoped().Where("user_id = ?", ownerID).Delete(&AuditEntry{}).Error
	})
}
