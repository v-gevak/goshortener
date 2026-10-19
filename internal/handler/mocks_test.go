package handler

type mockURLService struct {
	shorten func(originalURL string) string
	expand  func(key string) (string, bool)
}

func (m *mockURLService) Shorten(originalURL string) string {
	return m.shorten(originalURL)
}

func (m *mockURLService) Expand(key string) (string, bool) {
	return m.expand(key)
}
