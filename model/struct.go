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
	Title       string
	Url         string
	Cover       string
	Description string
	Authors     []string
	Chapters    []*Chapter
}

type Novel struct {
	Title       string
	Description string
	Authors     []string
	Volumes     []*Volume
}
