package model

type ExtraFile struct {
	Data         []byte
	Path         string
	ManifestItem ManifestItem
}

type Downloader interface {
	GetNovel(novelId int) (*Novel, error)
	GetVolume(novelId int, volumeId int) (*Volume, error)
	GetChapter(novelId int, volumeId int, chapterId int) (*Chapter, error)
	GetStyleCSS() string
	GetExtraFiles() []ExtraFile
}
