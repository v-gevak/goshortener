package model

type URLStore interface {
	Set(key string, originalURL string)
	Get(key string) (string, bool)
	Exists(url string) bool
}
