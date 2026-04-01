// Package storage - это пакет, который содержит логику взаимодействия с базой данных,
package storage

import "database/sql"

// Storage - это структура, которая будет использоваться для взаимодействия с базой данных.
type Storage struct {
	db *sql.DB
}

// New - это функция, которая инициализирует новый экземпляр Storage с заданной базой данных.
func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}
