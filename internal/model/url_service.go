package model

type URLService interface {
	BuildShortURL(originalURL string, host string) string
	Expand(key string) (string, bool)
}
