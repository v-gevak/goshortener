package handler

type mockURLService struct {
	shorten func(originalURL string) (string, error)
	expand  func(key string) (string, bool)
}

func (m *mockURLService) Shorten(originalURL string) (string, error) {
	return m.shorten(originalURL)
}

func (m *mockURLService) Expand(key string) (string, bool) {
	return m.expand(key)
}
