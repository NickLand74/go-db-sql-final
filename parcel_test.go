package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("failled to connect to database: %v", err)
	} // настройте подключение к БД
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	assert.NoError(t, err, "expected no error when adding parcel")
	assert.Positive(t, id, "expected positive id after addong pacel")

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	fetchedParcel, err := store.Get(id)
	assert.NoError(t, err, "expected no error when getting parcel")
	assert.Equal(t, parcel, fetchedParcel, "expected fetched parcel to match added parcel")

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	assert.NoError(t, err, "expected no error when deleting parcel")
	// Проверяем, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	assert.Error(t, err, "expected error when getting deleted parce")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close() // обязательно закройте соединение после теста

	store := NewParcelStore(db)
	originalParcel := getTestParcel() // получите тестовую посылку

	// add
	id, err := store.Add(originalParcel)
	require.NoError(t, err)
	require.NotZero(t, id, "Expected valid ID for the added parcel")

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Expected no error when updating address")

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err, "Expected no error when retrieving the updated parcel")

	// Проверяем, что адрес обновился
	assert.Equal(t, newAddress, updatedParcel.Address, "Address should be updated to the new address")
}

func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close() // обязательно закройте соединение после теста

	store := NewParcelStore(db)
	originalParcel := getTestParcel() // получите тестовую посылку

	// add
	id, err := store.Add(originalParcel)
	require.NoError(t, err) // убедитесь, что добавление прошло без ошибок
	require.NotZero(t, id, "Expected valid ID for the added parcel")

	// set status
	newStatus := "delivered"
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err, "Expected no error when updating status")

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err, "Expected no error when retrieving the updated parcel")

	// Проверяем, что статус обновился
	assert.Equal(t, newStatus, updatedParcel.Status, "Status should be updated to the new status")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close() // обязательно закройте соединение после теста
	// Создаем объект ParcelStore
	store := NewParcelStore(db)
	require.NotNil(t, store, "Expected to create a ParcelStore instance")

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := rand.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД
		require.NoError(t, err)          // убедитесь в отсутствии ошибки
		require.NotZero(t, id, "Expected valid ID for the added parcel")

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента
	require.NoError(t, err)                         // убедитесь в отсутствии ошибки
	require.Equal(t, len(parcels), len(storedParcels), "Expected number of parcels to match added parcels")

	// check
	for _, storedParcel := range storedParcels {
		originalParcel, exists := parcelMap[storedParcel.Number] // проверяем, что посылка существует в parcelMap
		require.True(t, exists, "Expected stored parcel in parcelMap")

		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, originalParcel.Client, storedParcel.Client, "Client ID should match")
		assert.Equal(t, originalParcel.Number, storedParcel.Number, "Parcel number should match")
		// Добавьте другие проверки полей, если необходимо
	}
}
