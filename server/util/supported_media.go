package util

// Types of media supported by Watcharr
// in an overarching way.
type SupportedMedia string

const (
	SupportedMediaMovie SupportedMedia = "movie"
	SupportedMediaShow  SupportedMedia = "tv"
	SupportedMediaGame  SupportedMedia = "game"
	SupportedMediaBook  SupportedMedia = "book"
)
