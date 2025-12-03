package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func openTestDB(t *testing.T) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}

	const createTableQuery = `
	CREATE TABLE IF NOT EXISTS parcel (
		number      INTEGER PRIMARY KEY AUTOINCREMENT,
		client      INTEGER NOT NULL,
		status      TEXT NOT NULL,
		address     TEXT NOT NULL,
		created_at  TEXT NOT NULL
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := openTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	parcel.Number = int(id)

	storedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)

	require.Equal(t, parcel, storedParcel)

	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	_, err = store.Get(parcel.Number)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db, err := openTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = int(id)

	newAddress := "new test address"

	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	parcel.Address = newAddress

	storedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)

	require.Equal(t, newAddress, storedParcel.Address)

	require.Equal(t, parcel.Status, storedParcel.Status)
}

func TestSetStatus(t *testing.T) {
	db, err := openTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = int(id)

	newStatus := ParcelStatusSent

	err = store.SetStatus(parcel.Number, newStatus)
	require.NoError(t, err)

	parcel.Status = newStatus

	storedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)

	require.Equal(t, newStatus, storedParcel.Status)

	require.Equal(t, parcel.Address, storedParcel.Address)
}

func TestGetByClient(t *testing.T) {
	db, err := openTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		parcels[i].Number = int(id)

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	require.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]

		require.True(t, ok)

		require.Equal(t, expectedParcel, parcel)
	}
}
