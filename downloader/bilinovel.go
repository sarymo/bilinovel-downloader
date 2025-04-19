package downloader

import (
	"bilinovel-downloader/model"
	"bilinovel-downloader/template"
	"bilinovel-downloader/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

func GetNovel(novelId int) (*model.Novel, error) {
	novelUrl := fmt.Sprintf("https://www.bilinovel.com/novel/%v.html", novelId)
	resp, err := utils.Request().Get(novelUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel info: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get novel info: %v", resp.Status())
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %v", err)
	}

	novel := &model.Novel{}

	novel.Title = strings.TrimSpace(doc.Find(".book-title").First().Text())
	novel.Description = strings.TrimSpace(doc.Find(".book-summary>content").First().Text())
	novel.Id = novelId

	doc.Find(".authorname>a").Each(func(i int, s *goquery.Selection) {
		novel.Authors = append(novel.Authors, strings.TrimSpace(s.Text()))
	})
	doc.Find(".illname>a").Each(func(i int, s *goquery.Selection) {
		novel.Authors = append(novel.Authors, strings.TrimSpace(s.Text()))
	})

	volumes, err := getNovelVolumes(novelId)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel volumes: %v", err)
	}
	novel.Volumes = volumes

	return novel, nil
}

func GetVolume(novelId int, volumeId int) (*model.Volume, error) {
	novelUrl := fmt.Sprintf("https://www.bilinovel.com/novel/%v/catalog", novelId)
	resp, err := utils.Request().Get(novelUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel info: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get novel info: %v", resp.Status())
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %v", err)
	}

	seriesIdx := 0
	doc.Find("a.volume-cover-img").Each(func(i int, s *goquery.Selection) {
		if s.AttrOr("href", "") == fmt.Sprintf("/novel/%v/vol_%v.html", novelId, volumeId) {
			seriesIdx = i + 1
		}
	})

	novelTitle := strings.TrimSpace(doc.Find(".book-title").First().Text())

	if seriesIdx == 0 {
		return nil, fmt.Errorf("volume not found: %v", volumeId)
	}

	volumeUrl := fmt.Sprintf("https://www.bilinovel.com/novel/%v/vol_%v.html", novelId, volumeId)
	resp, err = utils.Request().Get(volumeUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel info: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get novel info: %v", resp.Status())
	}

	doc, err = goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %v", err)
	}

	volume := &model.Volume{}
	volume.NovelId = novelId
	volume.NovelTitle = novelTitle
	volume.Id = volumeId
	volume.SeriesIdx = seriesIdx
	volume.Title = strings.TrimSpace(doc.Find(".book-title").First().Text())
	volume.Description = strings.TrimSpace(doc.Find(".book-summary>content").First().Text())
	volume.Cover = doc.Find(".book-cover").First().AttrOr("src", "")
	volume.Url = volumeUrl
	volume.Chapters = make([]*model.Chapter, 0)

	doc.Find(".authorname>a").Each(func(i int, s *goquery.Selection) {
		volume.Authors = append(volume.Authors, strings.TrimSpace(s.Text()))
	})
	doc.Find(".illname>a").Each(func(i int, s *goquery.Selection) {
		volume.Authors = append(volume.Authors, strings.TrimSpace(s.Text()))
	})

	doc.Find(".chapter-li.jsChapter").Each(func(i int, s *goquery.Selection) {
		volume.Chapters = append(volume.Chapters, &model.Chapter{
			Title: s.Find("a").Text(),
			Url:   fmt.Sprintf("https://www.bilinovel.com%v", s.Find("a").AttrOr("href", "")),
		})
	})

	return volume, nil
}

func getNovelVolumes(novelId int) ([]*model.Volume, error) {
	catelogUrl := fmt.Sprintf("https://www.bilinovel.com/novel/%v/catalog", novelId)
	resp, err := utils.Request().Get(catelogUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get catelog: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get catelog: %v", resp.Status())
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %v", err)
	}

	volumeRegexp := regexp.MustCompile(fmt.Sprintf(`/novel/%v/vol_(\d+).html`, novelId))

	volumeIds := make([]string, 0)
	doc.Find("a.volume-cover-img").Each(func(i int, s *goquery.Selection) {
		link := s.AttrOr("href", "")
		matches := volumeRegexp.FindStringSubmatch(link)
		if len(matches) > 0 {
			volumeIds = append(volumeIds, matches[1])
		}
	})

	volumes := make([]*model.Volume, 0)
	for i, volumeIdStr := range volumeIds {
		volumeId, err := strconv.Atoi(volumeIdStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert volume id: %v", err)
		}
		volume, err := GetVolume(novelId, volumeId)
		if err != nil {
			return nil, fmt.Errorf("failed to get volume info: %v", err)
		}
		volume.SeriesIdx = i
		volumes = append(volumes, volume)
	}

	return volumes, nil
}

func DownloadNovel(novelId int, outputPath string) error {
	log.Printf("Downloading Novel: %v", novelId)

	novel, err := GetNovel(novelId)
	if err != nil {
		return fmt.Errorf("failed to get novel info: %v", err)
	}

	outputPath = filepath.Join(outputPath, utils.CleanDirName(novel.Title))
	err = os.MkdirAll(outputPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	for _, volume := range novel.Volumes {
		err := downloadVolume(volume, outputPath)
		if err != nil {
			return fmt.Errorf("failed to download volume: %v", err)
		}
	}

	return nil
}

func DownloadVolume(novelId, volumeId int, outputPath string) error {
	volume, err := GetVolume(novelId, volumeId)
	if err != nil {
		return fmt.Errorf("failed to get volume info: %v", err)
	}
	err = downloadVolume(volume, outputPath)
	if err != nil {
		return fmt.Errorf("failed to download volume: %v", err)
	}
	return nil
}

func downloadVolume(volume *model.Volume, outputPath string) error {
	log.Printf("Downloading Volume: %s", volume.Title)
	outputPath = filepath.Join(outputPath, utils.CleanDirName(volume.Title))
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	_, err = os.Stat(filepath.Join(outputPath, "volume.json"))
	if os.IsNotExist(err) {
		for idx, chapter := range volume.Chapters {
			err := DownloadChapter(idx, chapter, outputPath)
			if err != nil {
				return fmt.Errorf("failed to download chapter: %v", err)
			}
		}
	} else {
		jsonBytes, err := os.ReadFile(filepath.Join(outputPath, "volume.json"))
		if err != nil {
			return fmt.Errorf("failed to read volume: %v", err)
		}
		err = json.Unmarshal(jsonBytes, volume)
		if err != nil {
			return fmt.Errorf("failed to unmarshal volume: %v", err)
		}
		for idx, chapter := range volume.Chapters {
			file, err := os.Create(filepath.Join(outputPath, fmt.Sprintf("OEBPS/Text/chapter-%03v.xhtml", idx+1)))
			if err != nil {
				return fmt.Errorf("failed to create chapter file: %v", err)
			}
			err = template.ContentXHTML(chapter).Render(context.Background(), file)
			if err != nil {
				return fmt.Errorf("failed to render text file: %v", err)
			}
		}
	}

	for i := range volume.Chapters {
		volume.Chapters[i].ImageFullPaths = utils.Unique(volume.Chapters[i].ImageFullPaths)
		volume.Chapters[i].ImageOEBPSPaths = utils.Unique(volume.Chapters[i].ImageOEBPSPaths)
	}

	jsonBytes, err := json.Marshal(volume)
	if err != nil {
		return fmt.Errorf("failed to marshal volume: %v", err)
	}
	err = os.WriteFile(filepath.Join(outputPath, "volume.json"), jsonBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write volume: %v", err)
	}

	coverPath := filepath.Join(outputPath, "OEBPS/Images/cover.jpg")
	err = os.MkdirAll(path.Dir(coverPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create cover directory: %v", err)
	}
	err = DownloadImg(volume.Cover, coverPath)
	if err != nil {
		return fmt.Errorf("failed to download cover: %v", err)
	}

	coverXHTMLPath := filepath.Join(outputPath, "OEBPS/Text/cover.xhtml")
	err = os.MkdirAll(path.Dir(coverXHTMLPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create cover directory: %v", err)
	}
	file, err := os.Create(coverXHTMLPath)
	if err != nil {
		return fmt.Errorf("failed to create cover file: %v", err)
	}
	err = template.ContentXHTML(&model.Chapter{
		Title:   "封面",
		Content: fmt.Sprintf(`<img src="../Images/cover%s" />`, path.Ext(volume.Cover)),
	}).Render(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to render cover: %v", err)
	}

	contentsXHTMLPath := filepath.Join(outputPath, "OEBPS/Text/contents.xhtml")
	err = os.MkdirAll(path.Dir(contentsXHTMLPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create contents directory: %v", err)
	}
	file, err = os.Create(contentsXHTMLPath)
	if err != nil {
		return fmt.Errorf("failed to create contents file: %v", err)
	}
	contents := strings.Builder{}
	contents.WriteString(`<nav epub:type="toc" id="toc">`)
	contents.WriteString(`<ol>`)
	for _, chapter := range volume.Chapters {
		contents.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`, strings.TrimPrefix(chapter.TextOEBPSPath, "Text/"), chapter.Title))
	}
	contents.WriteString(`</ol>`)
	contents.WriteString(`</nav>`)
	err = template.ContentXHTML(&model.Chapter{
		Title:   "目录",
		Content: contents.String(),
	}).Render(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to render contents: %v", err)
	}

	err = CreateContainerXML(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create container xml: %v", err)
	}

	u, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate uuid: %v", err)
	}

	err = CreateContentOPF(outputPath, u.String(), volume)
	if err != nil {
		return fmt.Errorf("failed to create content opf: %v", err)
	}

	err = CreateTocNCX(outputPath, u.String(), volume)
	if err != nil {
		return fmt.Errorf("failed to create toc ncx: %v", err)
	}

	err = utils.CreateEpub(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create epub: %v", err)
	}

	return nil
}

func DownloadChapter(chapterIdx int, chapter *model.Chapter, outputPath string) error {
	chapter.TextFullPath = filepath.Join(outputPath, fmt.Sprintf("OEBPS/Text/chapter-%03v.xhtml", chapterIdx+1))
	chapter.TextOEBPSPath = fmt.Sprintf("Text/chapter-%03v.xhtml", chapterIdx+1)
	err := os.MkdirAll(path.Dir(chapter.TextFullPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create text directory: %v", err)
	}

	page := 1
	for {
		hasNext, err := downloadChapterByPage(page, chapterIdx, chapter, outputPath)
		if err != nil {
			return fmt.Errorf("failed to download chapter: %v", err)
		}
		if !hasNext {
			break
		}
		page++
		time.Sleep(time.Second)
	}

	file, err := os.Create(chapter.TextFullPath)
	if err != nil {
		return fmt.Errorf("failed to create text file: %v", err)
	}

	err = template.ContentXHTML(chapter).Render(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to render text file: %v", err)
	}

	return nil
}

func downloadChapterByPage(page, chapterIdx int, chapter *model.Chapter, outputPath string) (bool, error) {
	Url := strings.TrimSuffix(chapter.Url, ".html") + fmt.Sprintf("_%v.html", page)
	log.Printf("Downloading Chapter: %s", Url)

	hasNext := false
	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language": "zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7,zh-TW;q=0.6",
		"Cookie":          "night=1;",
	}
	resp, err := utils.Request().SetHeaders(headers).Get(Url)
	if err != nil {
		return hasNext, err
	}
	if resp.StatusCode() != http.StatusOK {
		return hasNext, fmt.Errorf("failed to get chapter: %v", resp.Status())
	}

	if strings.Contains(resp.String(), `<a onclick="window.location.href = ReadParams.url_next;">下一頁</a>`) {
		hasNext = true
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		fmt.Println(err)
		return hasNext, err
	}

	imgSavePath := fmt.Sprintf("OEBPS/Images/chapter-%03v", chapterIdx+1)

	content := doc.Find("#acontent").First()
	content.Find(".cgo").Remove()
	content.Find("center").Remove()
	content.Find(".google-auto-placed").Remove()

	content.Find("img").Each(func(i int, s *goquery.Selection) {
		if err != nil {
			return
		}
		imgUrl := s.AttrOr("data-src", "")
		if imgUrl == "" {
			imgUrl = s.AttrOr("src", "")
			if imgUrl == "" {
				return
			}
		}

		fileName := filepath.Join(imgSavePath, fmt.Sprintf("%03v%s", i+1, path.Ext(imgUrl)))
		err = DownloadImg(imgUrl, filepath.Join(outputPath, fileName))
		if err == nil {
			s.SetAttr("src", "../"+strings.TrimPrefix(fileName, "OEBPS/"))
			s.RemoveAttr("class")
			s.RemoveAttr("data-src")
			if s.AttrOr("alt", "") == "" {
				s.SetAttr("alt", fmt.Sprintf("image-%03d", i+1))
			}
			chapter.ImageFullPaths = append(chapter.ImageFullPaths, filepath.Join(outputPath, fileName))
			chapter.ImageOEBPSPaths = append(chapter.ImageOEBPSPaths, strings.TrimPrefix(fileName, "OEBPS/"))
		}
	})
	if err != nil {
		return false, fmt.Errorf("failed to download img: %v", err)
	}

	html, err := content.Html()
	if err != nil {
		return false, fmt.Errorf("failed to get html: %v", err)
	}

	chapter.Content += strings.TrimSpace(html)

	return hasNext, nil
}

func DownloadImg(url string, fileName string) error {
	_, err := os.Stat(fileName)
	if !os.IsNotExist(err) {
		return nil
	}

	log.Printf("Downloading Image: %s", url)
	dir := filepath.Dir(fileName)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	resp, err := utils.Request().SetHeader("Referer", "https://www.bilinovel.com").Get(url)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, resp.Body(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func CreateContainerXML(dirPath string) error {
	containerPath := filepath.Join(dirPath, "META-INF/container.xml")
	err := os.MkdirAll(path.Dir(containerPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create container directory: %v", err)
	}
	file, err := os.Create(containerPath)
	if err != nil {
		return fmt.Errorf("failed to create container file: %v", err)
	}
	err = template.ContainerXML().Render(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to render container: %v", err)
	}
	return nil
}

func CreateContentOPF(dirPath string, uuid string, volume *model.Volume) error {
	creators := make([]model.DCCreator, 0)
	for _, author := range volume.Authors {
		creators = append(creators, model.DCCreator{
			Value: author,
		})
	}
	dc := &model.DublinCoreMetadata{
		Titles: []model.DCTitle{
			{
				Value: volume.Title,
			},
		},
		Identifiers: []model.DCIdentifier{
			{
				Value: fmt.Sprintf("urn:uuid:%s", uuid),
				ID:    "book-id",
				// Scheme: "UUID",
			},
		},
		Languages: []model.DCLanguage{
			{
				Value: "zh-TW",
			},
		},
		Descriptions: []model.DCDescription{
			{
				Value: volume.Description,
			},
		},
		Creators: creators,
		Metas: []model.DublinCoreMeta{
			{
				Name:    "cover",
				Content: fmt.Sprintf("Images/cover%s", path.Ext(volume.Cover)),
			},
			{
				Property: "dcterms:modified",
				Value:    time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			},
			{
				Name:    "calibre:series",
				Content: volume.NovelTitle,
			},
			{
				Name:    "calibre:series_index",
				Content: strconv.Itoa(volume.SeriesIdx),
			},
		},
	}
	manifest := &model.Manifest{
		Items: make([]model.ManifestItem, 0),
	}
	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID:    "ncx",
		Link:  "toc.ncx",
		Media: "application/x-dtbncx+xml",
	})
	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID:    "cover",
		Link:  "Text/cover.xhtml",
		Media: "application/xhtml+xml",
	})
	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID:         "contents",
		Link:       "Text/contents.xhtml",
		Media:      "application/xhtml+xml",
		Properties: "nav",
	})
	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID:    "images-cover",
		Link:  fmt.Sprintf("Images/cover%s", path.Ext(volume.Cover)),
		Media: fmt.Sprintf("image/%s", strings.ReplaceAll(strings.TrimPrefix(path.Ext(volume.Cover), "."), "jpg", "jpeg")),
	})
	for _, chapter := range volume.Chapters {
		manifest.Items = append(manifest.Items, model.ManifestItem{
			ID:    path.Base(chapter.TextOEBPSPath),
			Link:  chapter.TextOEBPSPath,
			Media: "application/xhtml+xml",
		})
		for _, image := range chapter.ImageOEBPSPaths {
			item := model.ManifestItem{
				ID:   strings.Join(strings.Split(strings.ToLower(image), string(filepath.Separator)), "-"),
				Link: image,
			}
			item.Media = fmt.Sprintf("image/%s", strings.ReplaceAll(strings.TrimPrefix(path.Ext(volume.Cover), "."), "jpg", "jpeg"))
			manifest.Items = append(manifest.Items, item)
		}
	}
	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID:    "style",
		Link:  "Styles/style.css",
		Media: "text/css",
	})

	spine := &model.Spine{
		Items: make([]model.SpineItem, 0),
	}
	for _, item := range manifest.Items {
		if filepath.Ext(item.Link) == ".xhtml" {
			spine.Items = append(spine.Items, model.SpineItem{
				IDref: item.ID,
			})
		}
	}
	contentOPFPath := filepath.Join(dirPath, "OEBPS/content.opf")
	err := os.MkdirAll(path.Dir(contentOPFPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create content directory: %v", err)
	}
	file, err := os.Create(contentOPFPath)
	if err != nil {
		return fmt.Errorf("failed to create content file: %v", err)
	}
	err = template.ContentOPF("book-id", dc, manifest, spine, nil).Render(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to render content: %v", err)
	}
	return nil
}

func CreateTocNCX(dirPath string, uuid string, volume *model.Volume) error {
	navMap := &model.NavMap{Points: make([]*model.NavPoint, 0)}
	navMap.Points = append(navMap.Points, &model.NavPoint{
		Id:        "cover",
		PlayOrder: 1,
		Label:     "封面",
		Content:   model.NavPointContent{Src: "Text/cover.xhtml"},
	})
	navMap.Points = append(navMap.Points, &model.NavPoint{
		Id:        "contents",
		PlayOrder: 2,
		Label:     "目录",
		Content:   model.NavPointContent{Src: "Text/contents.xhtml"},
	})
	for idx, chapter := range volume.Chapters {
		navMap.Points = append(navMap.Points, &model.NavPoint{
			Id:        fmt.Sprintf("chapter-%03v", idx+1),
			PlayOrder: len(navMap.Points) + 1,
			Label:     chapter.Title,
			Content:   model.NavPointContent{Src: chapter.TextOEBPSPath},
		})
	}

	head := &model.TocNCXHead{
		Meta: []model.TocNCXHeadMeta{
			{Name: "dtb:uid", Content: fmt.Sprintf("urn:uuid:%s", uuid)},
		},
	}

	ncxPath := filepath.Join(dirPath, "OEBPS/toc.ncx")
	err := os.MkdirAll(path.Dir(ncxPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create toc directory: %v", err)
	}
	file, err := os.Create(ncxPath)
	if err != nil {
		return fmt.Errorf("failed to create toc file: %v", err)
	}
	err = template.TocNCX(volume.Title, head, navMap).Render(context.Background(), file)
	if err != nil {
		return fmt.Errorf("failed to render toc: %v", err)
	}
	return nil
}
