package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/golangTroshin/shorturl/internal/app/http/middleware"
)

// FileStore represents the file-based storage for URLs.
// It uses in-memory synchronization with a file to persist and retrieve URL data.
type FileStore struct {
	mu      sync.RWMutex
	urlList map[string]URL
}

// NewFileStore initializes and returns a new FileStore instance.
// It loads existing data from the file specified in the configuration.
func NewFileStore() (*FileStore, error) {
	store := &FileStore{
		urlList: make(map[string]URL),
	}

	err := store.loadFromFile()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Get retrieves the original URL corresponding to a short URL.
// Returns an error if the short URL does not exist in the store.
func (store *FileStore) Get(ctx context.Context, key string) (string, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	val, ok := store.urlList[key]
	if !ok {
		return "", errors.New("no info about requested route")
	}

	return val.OriginalURL, nil
}

// GetByUserID retrieves all URLs associated with a given user ID.
// Currently, the implementation does not return any data (placeholder).
func (store *FileStore) GetByUserID(_ context.Context, userID string) ([]URL, error) {
	var URLs []URL

	return URLs, nil
}

// Set adds a new URL to the store, generating a unique short URL for it.
// The URL is written to the file for persistence.
func (store *FileStore) Set(ctx context.Context, value string) (URL, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	userID := ctx.Value(middleware.UserIDKey).(string)
	url := getURLObject(value, userID)
	store.urlList[url.ShortURL] = url

	Producer, err := NewProducer(config.Options.StoragePath)
	if err != nil {
		return url, err
	}
	defer Producer.Close()

	if err := Producer.WriteURL(&url); err != nil {
		return url, err
	}

	return url, nil
}

// SetBatch adds multiple URLs to the store in a single operation.
// Each URL is persisted to the file.
func (store *FileStore) SetBatch(ctx context.Context, batch []RequestBodyBanch) ([]URL, error) {
	URLs := make([]URL, 0, len(batch))

	store.mu.Lock()
	defer store.mu.Unlock()

	userID := ctx.Value(middleware.UserIDKey).(string)
	for _, url := range batch {
		urlObj := getURLObjectWithID(url.CorrelationID, url.OriginalURL, userID)
		store.urlList[urlObj.ShortURL] = urlObj

		Producer, err := NewProducer(config.Options.StoragePath)
		if err != nil {
			return URLs, err
		}
		defer Producer.Close()

		if err := Producer.WriteURL(&urlObj); err != nil {
			return URLs, err
		}

		URLs = append(URLs, urlObj)
	}

	return URLs, nil
}

// BatchDeleteURLs marks multiple URLs as deleted for a specific user ID.
// Updates are made in memory; persistence is not yet implemented.
func (store *FileStore) BatchDeleteURLs(userID string, batch []string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.urlList) == 0 {
		return errors.New("no URLs in the store")
	}

	batchMap := make(map[string]struct{})
	for _, shortURL := range batch {
		batchMap[shortURL] = struct{}{}
	}

	for key, url := range store.urlList {
		if url.UserID == userID {
			if _, found := batchMap[url.ShortURL]; found {
				url.DeletedFlag = true
				store.urlList[key] = url
			}
		}
	}

	return nil
}

// loadFromFile loads URL data from the file into the in-memory store.
func (store *FileStore) loadFromFile() error {
	consumer, err := NewConsumer(config.Options.StoragePath)
	if err != nil {
		return err
	}
	defer consumer.Close()

	for {
		url, err := consumer.ReadURL()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		if _, ok := store.urlList[url.ShortURL]; !ok {
			store.urlList[url.ShortURL] = *url
		}
	}
	return nil
}

// Producer is responsible for writing URL data to the file in JSON format.
type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

// NewProducer creates a new Producer for writing to the specified file path.
func NewProducer(filePath string) (*Producer, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// Close closes the file handle for the Producer.
func (p *Producer) Close() error {
	return p.file.Close()
}

// WriteURL writes a URL object to the file in JSON format, appending a newline.
func (p *Producer) WriteURL(url *URL) error {
	data, err := json.Marshal(&url)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

// Consumer is responsible for reading URL data from the file in JSON format.
type Consumer struct {
	file   *os.File
	reader *bufio.Reader
}

// NewConsumer creates a new Consumer for reading from the specified file path.
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

// ReadURL reads and unmarshals a URL object from the file.
func (c *Consumer) ReadURL() (*URL, error) {
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	url := URL{}
	err = json.Unmarshal(data, &url)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

// Close closes the file handle for the Consumer.
func (c *Consumer) Close() error {
	return c.file.Close()
}

// GetStats retrieves service statistic
func (store *FileStore) GetStats(_ context.Context) (Stats, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var stats Stats

	stats.Urls = len(store.urlList)

	userIDSet := make(map[string]struct{})

	for _, url := range store.urlList {
		userIDSet[url.UserID] = struct{}{}
	}

	stats.Users = len(userIDSet)

	return stats, nil
}
