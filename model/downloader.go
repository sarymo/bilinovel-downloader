package model

type ExtraFile struct {
	Data         []byte
	Path         string
	ManifestItem ManifestItem
}

type Downloader interface {
	GetNovel(novelId int, skipChapter bool) (*Novel, error)
	GetVolume(novelId int, volumeId int, skipChapter bool) (*Volume, error)
	GetChapter(novelId int, volumeId int, chapterId int) (*Chapter, error)
	GetStyleCSS() string
	GetExtraFiles() []ExtraFile
	Close() error
}
