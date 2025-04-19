package model

type Chapter struct {
	Title           string
	Url             string
	Content         string
	ImageOEBPSPaths []string
	ImageFullPaths  []string
	TextOEBPSPath   string
	TextFullPath    string
}

type Volume struct {
	Id          int
	SeriesIdx   int
	Title       string
	Url         string
	Cover       string
	Description string
	Authors     []string
	Chapters    []*Chapter
	NovelId     int
	NovelTitle  string
}

type Novel struct {
	Id          int
	Title       string
	Description string
	Authors     []string
	Volumes     []*Volume
}
