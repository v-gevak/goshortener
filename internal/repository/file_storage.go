package repository

import (
	"bufio"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/base62"
	"github.com/v-gevak/goshortener/internal/entity"
)

type fileStorage struct {
	file       *os.File
	filePath   string
	fileReader *bufio.Reader
	fileWriter *jsontext.Encoder
	lastID     int64
	mux        sync.Mutex
}

func newFileStorage(filePath string) (*fileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if err != nil {
		return nil, err
	}

	store := &fileStorage{
		filePath:   filePath,
		fileReader: bufio.NewReader(file),
		fileWriter: jsontext.NewEncoder(file),
		file:       file,
	}
	if err := store.restoreLastID(); err != nil {
		file.Close()
		return nil, err
	}

	return store, nil
}

type Item struct {
	ID          int64  `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (p *fileStorage) restoreLastID() error {
	for {
		line, err := p.fileReader.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if len(line) == 0 {
			return nil
		}

		var item Item
		if err := json.Unmarshal(line, &item); err != nil {
			return err
		}
		p.lastID = item.ID

		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

func (p *fileStorage) Get(key string) (*entity.URL, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	item, err := p.findItem(func(item Item) bool {
		return item.ShortURL == key
	})

	if err != nil {
		return nil, err
	}

	return &entity.URL{Key: item.ShortURL, Original: item.OriginalURL}, nil
}

func (p *fileStorage) findItem(match func(item Item) bool) (*Item, error) {
	if _, err := p.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	p.fileReader.Reset(p.file)

	for {
		line, err := p.fileReader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, apperrors.ErrNotFound
			}

			return nil, err
		}

		var item Item

		if err := json.Unmarshal(line, &item); err != nil {
			return nil, err
		}

		if match(item) {
			return &item, nil
		}
	}
}

func (p *fileStorage) SetIfAbsent(originalURL string) (*entity.URL, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	item, err := p.findItem(func(item Item) bool {
		return item.OriginalURL == originalURL
	})

	if err == nil {
		return &entity.URL{Key: item.ShortURL, Original: item.OriginalURL}, nil
	}

	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	id := p.lastID + 1

	item = &Item{
		ID:          id,
		ShortURL:    base62.EncodeID(id),
		OriginalURL: originalURL,
	}

	if err := json.MarshalEncode(p.fileWriter, item); err != nil {
		return nil, err
	}
	p.lastID = id

	return &entity.URL{Key: item.ShortURL, Original: item.OriginalURL}, nil
}

func (p *fileStorage) Ping() error {
	return nil
}

func (p *fileStorage) Close() error {
	return p.file.Close()
}
