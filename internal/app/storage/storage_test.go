package storage

import (
	"testing"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func TestGetStorageByConfig_MemoryStore(t *testing.T) {
	config.Options.DatabaseDsn = ""
	config.Options.StoragePath = ""

	store, err := GetStorageByConfig()

	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestGetStorageByConfig_FileStore(t *testing.T) {
	config.Options.DatabaseDsn = ""
	config.Options.StoragePath = "test_storage.json"

	store, err := GetStorageByConfig()

	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.FileExists(t, config.Options.StoragePath)
}

func TestGenerateShortURL(t *testing.T) {
	input := "https://example.com"
	expectedLength := 8

	shortURL := generateShortURL(input)

	assert.Equal(t, expectedLength, len(shortURL))
}

func TestGetURLObject(t *testing.T) {
	url := "https://example.com"
	userID := "test-user"

	urlObject := getURLObject(url, userID)

	assert.Equal(t, "uuid_"+generateShortURL(url), urlObject.UUID)
	assert.Equal(t, generateShortURL(url), urlObject.ShortURL)
	assert.Equal(t, url, urlObject.OriginalURL)
	assert.Equal(t, userID, urlObject.UserID)
}

func TestGetURLObjectWithID(t *testing.T) {
	uuid := "test-uuid"
	url := "https://example.com"
	userID := "test-user"

	urlObject := getURLObjectWithID(uuid, url, userID)

	assert.Equal(t, uuid, urlObject.UUID)
	assert.Equal(t, generateShortURL(url), urlObject.ShortURL)
	assert.Equal(t, url, urlObject.OriginalURL)
	assert.Equal(t, userID, urlObject.UserID)
}
