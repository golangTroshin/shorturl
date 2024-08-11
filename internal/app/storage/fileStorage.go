package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/golangTroshin/shorturl/internal/app/config"
)

type FileStore struct {
	mu      sync.RWMutex
	urlList map[string]URL
}

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

func (store *FileStore) Get(ctx context.Context, key string) (string, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	val, ok := store.urlList[key]
	if !ok {
		return "", errors.New("no info about requested route")
	}

	return val.OriginalURL, nil
}

func (store *FileStore) Set(ctx context.Context, value []byte) (URL, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	key := generateKey(value)
	url := URL{
		UUID:        len(store.urlList) + 1,
		ShortURL:    key,
		OriginalURL: string(value),
	}
	store.urlList[key] = url

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

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

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

func (p *Producer) Close() error {
	return p.file.Close()
}

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

type Consumer struct {
	file   *os.File
	reader *bufio.Reader
}

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

func (c *Consumer) Close() error {
	return c.file.Close()
}
