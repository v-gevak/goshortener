package model

type URLService interface {
	Shorten(originalURL string) string
	Expand(key string) (string, bool)
}
