package repository

import (
	"bufio"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"sync"
	"uuid"

	"github.com/v-gevak/goshortener/internal/apperrors"
	"github.com/v-gevak/goshortener/internal/urlutil"
)

type fileProvider struct {
	file       *os.File
	filePath   string
	fileReader *bufio.Reader
	fileWriter *jsontext.Encoder
	mux        sync.Mutex
}

func newFileProvider(filePath string) (*fileProvider, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if err != nil {
		return nil, err
	}

	return &fileProvider{
		filePath:   filePath,
		fileReader: bufio.NewReader(file),
		fileWriter: jsontext.NewEncoder(file),
		file:       file,
	}, err
}

type Item struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

func (p *fileProvider) Get(key string) (string, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	item, err := p.findItem(func(item Item) (bool, error) {
		return item.ShortURL == key, nil
	})

	if err != nil {
		return "", err
	}

	return item.OriginalURL, nil
}

func (p *fileProvider) findItem(match func(item Item) (bool, error)) (*Item, error) {
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

		matches, err := match(item)
		if err != nil {
			return nil, err
		}
		if matches {
			return &item, nil
		}
	}
}

func (p *fileProvider) SetIfAbsent(key, originalURL string) (string, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	keyExists := false

	item, err := p.findItem(func(item Item) (bool, error) {
		if item.ShortURL == key {
			keyExists = true
		}

		return urlutil.CompareURLs(item.OriginalURL, originalURL)
	})

	if err == nil {
		return item.ShortURL, nil
	}

	if !errors.Is(err, apperrors.ErrNotFound) {
		return "", err
	}

	if keyExists {
		return "", apperrors.ErrConflict
	}

	item = &Item{
		UUID:        uuid.New(),
		ShortURL:    key,
		OriginalURL: originalURL,
	}

	if err := json.MarshalEncode(p.fileWriter, item); err != nil {
		return "", err
	}

	return item.ShortURL, nil
}

func (p *fileProvider) Close() error {
	return p.file.Close()
}
