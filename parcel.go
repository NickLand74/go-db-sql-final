package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore создает новый экземпляр ParcelStore.
func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{db: db}
}

// Add добавляет новый пакет в базу данных и возвращает ID нового пакета или ошибку.
func (store *ParcelStore) Add(parcel Parcel) (int, error) {
	result, err := store.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		parcel.Client, parcel.Status, parcel.Address, parcel.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get получает пакет по его номеру и возвращает ошибку, если он не найден.
func (s *ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	query := "SELECT number, client, address, status, created_at FROM parcel WHERE number = ?"
	err := s.db.QueryRow(query, number).Scan(&p.Number, &p.Client, &p.Address, &p.Status, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, errors.New("пакет не найден")
		}
		return Parcel{}, err
	}
	return p, nil
}

// GetByClient получает все пакеты, связанные с конкретным клиентом.
func (s *ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel
	query := "SELECT number, client, address, status, created_at FROM parcel WHERE client = ?"
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Address, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// SetStatus обновляет статус пакета по его номеру.
func (s *ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel SET status = ? WHERE number = ?"
	_, err := s.db.Exec(query, status, number)
	return err
}

// SetAddress изменяет адрес пакета, если его статус 'registered'.
func (s *ParcelStore) SetAddress(number int, address string) error {
	// Сначала проверяем статус посылки
	var status string
	query := "SELECT status FROM parcel WHERE number = ?"
	err := s.db.QueryRow(query, number).Scan(&status)
	if err != nil {
		return err
	}

	if status != "registered" {
		return errors.New("адрес может быть изменён только если статус 'registered'")
	}

	// Изменяем адрес только если статус 'registered'
	query = "UPDATE parcel SET address = ? WHERE number = ?"
	_, err = s.db.Exec(query, address, number)
	return err
}

// Delete удаляет пакет, если его статус 'registered'.
func (s *ParcelStore) Delete(number int) error {
	// Сначала проверяем статус посылки
	var status string
	query := "SELECT status FROM parcel WHERE number = ?"
	err := s.db.QueryRow(query, number).Scan(&status)
	if err != nil {
		return err
	}

	if status != "registered" {
		return errors.New("можно удалять только пакеты со статусом 'registered'")
	}

	// Удаляем строку если статус 'registered'
	query = "DELETE FROM parcel WHERE number = ?"
	_, err = s.db.Exec(query, number)
	return err
}
