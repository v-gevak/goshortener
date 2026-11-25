package repository

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
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
		file:       file,
	}
	if err := store.restoreLastID(); err != nil {
		file.Close()
		return nil, err
	}
	store.lastID = max(store.lastID, firstShortID-1)

	return store, nil
}

type Item struct {
	ID          int64  `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
}

func (s *fileStorage) restoreLastID() error {
	for {
		line, err := s.fileReader.ReadBytes('\n')
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
		s.lastID = item.ID

		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

func (s *fileStorage) Get(_ context.Context, key string) (*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	item, err := s.findItem(func(item Item) bool {
		return item.ShortURL == key
	})

	if err != nil {
		return nil, err
	}

	return &entity.URL{Key: item.ShortURL, Original: item.OriginalURL, UserID: item.UserID}, nil
}

func (s *fileStorage) GetBatchByUser(ctx context.Context, userID string) ([]*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	s.fileReader.Reset(s.file)

	var entities []*entity.URL

	for {
		line, err := s.fileReader.ReadBytes('\n')
		isEOF := errors.Is(err, io.EOF)
		if err != nil && !isEOF {
			return nil, err
		}

		if isEOF && len(line) == 0 {
			break
		}

		var item Item

		if err := json.Unmarshal(line, &item); err != nil {
			return nil, err
		}

		if item.UserID == userID {
			entities = append(
				entities,
				&entity.URL{Key: item.ShortURL, Original: item.OriginalURL, UserID: item.UserID},
			)
		}

		if isEOF {
			break
		}
	}

	return entities, nil
}

func (s *fileStorage) findItem(match func(item Item) bool) (*Item, error) {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	s.fileReader.Reset(s.file)

	for {
		line, err := s.fileReader.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if len(line) == 0 {
			return nil, apperrors.ErrNotFound
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

func (s *fileStorage) SetIfAbsent(_ context.Context, originalURL, userID string) (*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	item, err := s.findItem(func(item Item) bool {
		return item.OriginalURL == originalURL
	})

	if err == nil {
		return &entity.URL{Key: item.ShortURL, Original: item.OriginalURL, UserID: item.UserID}, apperrors.ErrConflict
	}

	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	id := s.lastID + 1

	item = &Item{
		ID:          id,
		ShortURL:    base62.EncodeID(id),
		OriginalURL: originalURL,
		UserID:      userID,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	if err := appendRecords(s.file, append(data, '\n')); err != nil {
		return nil, err
	}
	s.lastID = id

	return &entity.URL{Key: item.ShortURL, Original: item.OriginalURL, UserID: item.UserID}, nil
}

func (s *fileStorage) SetBatchIfAbsent(_ context.Context, originalURLs []string, userID string) ([]*entity.URL, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	result := make([]*entity.URL, len(originalURLs))

	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	s.fileReader.Reset(s.file)
	existing := make(map[string]Item)
	for {
		line, err := s.fileReader.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if len(line) > 0 {
			var item Item
			if err := json.Unmarshal(line, &item); err != nil {
				return nil, err
			}
			existing[item.OriginalURL] = item
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}

	var buffer bytes.Buffer
	encoder := jsontext.NewEncoder(&buffer)
	nextID := s.lastID
	for i, originalURL := range originalURLs {
		item, exists := existing[originalURL]
		if !exists {
			nextID++
			item = Item{ID: nextID, ShortURL: base62.EncodeID(nextID), OriginalURL: originalURL, UserID: userID}
			if err := json.MarshalEncode(encoder, item); err != nil {
				return nil, err
			}
			existing[originalURL] = item
		}
		result[i] = &entity.URL{Key: item.ShortURL, Original: item.OriginalURL, UserID: item.UserID}
	}

	if buffer.Len() == 0 {
		return result, nil
	}

	if err := appendRecords(s.file, buffer.Bytes()); err != nil {
		return nil, err
	}

	s.lastID = nextID
	return result, nil
}

func appendRecords(file *os.File, data []byte) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	fileSizeBeforeWrite := info.Size()
	if fileSizeBeforeWrite > 0 {
		var lastByte [1]byte
		if _, err := file.ReadAt(lastByte[:], fileSizeBeforeWrite-1); err != nil {
			return err
		}
		if lastByte[0] != '\n' {
			data = append([]byte{'\n'}, data...)
		}
	}

	n, err := file.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}

	if err != nil {
		if rollbackErr := file.Truncate(fileSizeBeforeWrite); rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("rollback append: %w", rollbackErr))
		}
		return err
	}

	return nil
}

func (s *fileStorage) Ping(context.Context) error {
	return nil
}

func (s *fileStorage) Close() error {
	return s.file.Close()
}
