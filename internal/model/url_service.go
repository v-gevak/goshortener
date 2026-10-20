package model

type URLService interface {
	Shorten(originalURL string) (string, error)
	Expand(key string) (string, bool)
}
