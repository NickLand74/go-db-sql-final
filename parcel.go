package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s *ParcelStore) Add(p Parcel) (int, error) {
	var lastInsertID int
	query := "INSERT INTO parcel (client, address, status) VALUES ($1, $2, $3) RETURNING number"
	err := s.db.QueryRow(query, p.Client, p.Address, p.Status).Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}
	return lastInsertID, nil
}

func (s *ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	query := "SELECT number, client, address, status FROM parcel WHERE number = $1"
	err := s.db.QueryRow(query, number).Scan(&p.Number, &p.Client, &p.Address, &p.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, errors.New("parcel not found")
		}
		return Parcel{}, err
	}
	return p, nil
}

func (s *ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel
	query := "SELECT number, client, address, status FROM parcel WHERE client = $1"
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Address, &p.Status); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (s *ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel SET status = $1 WHERE number = $2"
	_, err := s.db.Exec(query, status, number)
	return err
}

func (s *ParcelStore) SetAddress(number int, address string) error {
	// Сначала проверяем статус посылки
	var status string
	query := "SELECT status FROM parcel WHERE number = $1"
	err := s.db.QueryRow(query, number).Scan(&status)
	if err != nil {
		return err
	}

	if status != "registered" {
		return errors.New("address can only be changed if status is 'registered'")
	}

	// Изменяем адрес только если статус 'registered'
	query = "UPDATE parcel SET address = $1 WHERE number = $2"
	_, err = s.db.Exec(query, address, number)
	return err
}

func (s *ParcelStore) Delete(number int) error {
	// Сначала проверяем статус посылки
	var status string
	query := "SELECT status FROM parcel WHERE number = $1"
	err := s.db.QueryRow(query, number).Scan(&status)
	if err != nil {
		return err
	}

	if status != "registered" {
		return errors.New("only parcels with status 'registered' can be deleted")
	}

	// Удаляем строку если статус 'registered'
	query = "DELETE FROM parcel WHERE number = $1"
	_, err = s.db.Exec(query, number)
	return err
}
