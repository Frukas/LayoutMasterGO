package repository

import "gorm.io/gorm"

// Paginate returns a GORM Scope that handles page offsets and safety limits
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Fallback rules for invalid page numbers
		if page <= 0 {
			page = 1
		}

		// Security limit to prevent huge queries (max 100 items per page)
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10 // Default page size
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
