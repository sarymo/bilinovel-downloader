package model

type ChaperContent struct {
	Html   string
	Images map[string][]byte
}

type Chapter struct {
	Id       int
	NovelId  int
	VolumeId int
	Title    string
	Url      string
	Content  *ChaperContent
}

type Volume struct {
	Id          int
	SeriesIdx   int
	Title       string
	Url         string
	CoverUrl    string
	Cover       []byte
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
