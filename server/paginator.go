package main

import (
	"gorm.io/gorm"
)

// Parameters that the paginator uses to
// know what to return.
type PaginationParams struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
	// TODO sorting and filtering to be params?
	// Have to think about if this is better in a reusable
	// fashion (eg query params target db cols) or not.
	Sort string `json:"sort"`
}

// Pagination response struct.
type PaginationResponse struct {
	PaginationParams
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

// func paginate(value interface{}, pagination *pkg.Pagination, db *gorm.DB) *gorm.DB {
// 	var totalRows int64
// 	db.Model(value).Count(&totalRows)

// 	pagination.TotalRows = totalRows
// 	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
// 	pagination.TotalPages = totalPages

// 	offset := (page - 1) * pageSize
// 	return db.Offset(offset).Limit(pageSize)
// }

func Paginate(p PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.Page - 1) * p.Limit
		return db.Offset(offset).Limit(p.Limit)
	}
}
