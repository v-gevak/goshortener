package repository

type Store struct {
	urls map[string]string
}

func New() *Store {
	return &Store{urls: make(map[string]string)}
}

func (s *Store) Get(key string) (string, bool) {
	v, ok := s.urls[key]
	return v, ok
}

func (s *Store) Set(key string, originalURL string) {
	s.urls[key] = originalURL
}

func (s *Store) Exists(url string) bool {
	_, exists := s.urls[url]

	return exists
}
