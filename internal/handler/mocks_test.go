package handler

type mockURLService struct {
	buildShortURL func(originalURL string, host string) string
	expand        func(key string) (string, bool)
}

func (m *mockURLService) BuildShortURL(originalURL string, host string) string {
	return m.buildShortURL(originalURL, host)
}

func (m *mockURLService) Expand(key string) (string, bool) {
	return m.expand(key)
}
