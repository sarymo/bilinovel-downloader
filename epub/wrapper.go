package epub

import (
	"archive/zip"
	"bilinovel-downloader/model"
	"bilinovel-downloader/template"
	"bilinovel-downloader/utils"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func PackVolumeToEpub(volume *model.Volume, outputPath string, styleCSS string, extraFiles []model.ExtraFile) error {
	// 以“卷名”建立工作目录
	outputPath = filepath.Join(outputPath, utils.CleanDirName(volume.Title))
	if st, err := os.Stat(outputPath); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(outputPath, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
		} else {
			return fmt.Errorf("failed to stat output directory: %v", err)
		}
	} else if st.IsDir() {
		// 清空重建
		if err := os.RemoveAll(outputPath); err != nil {
			return fmt.Errorf("failed to remove output directory: %v", err)
		}
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("failed to re-create output directory: %v", err)
		}
	}

	// 先按规则选封面（仅设置 volume.Cover / volume.CoverUrl，不删除章节图片）
	chooseCover(volume)

	// 将文字写入 OEBPS/Text/chapter-%03v.xhtml
	// 将图片写入 OEBPS/Images/chapter-%03v/
	for i := range volume.Chapters {
		chapter := volume.Chapters[i] // *model.Chapter
		if chapter == nil {
			continue
		}

		// 写图片到该章目录
		imageNames := make([]string, 0, len(chapter.Content.Images))
		for imgName, imgData := range chapter.Content.Images {
			imageNames = append(imageNames, imgName)
			imgPath := filepath.Join(outputPath, fmt.Sprintf("OEBPS/Images/chapter-%03v/%s", i, imgName))
			if err := os.MkdirAll(filepath.Dir(imgPath), 0755); err != nil {
				return fmt.Errorf("failed to create image directory: %v", err)
			}
			if err := os.WriteFile(imgPath, imgData, 0644); err != nil {
				return fmt.Errorf("failed to write image: %v", err)
			}
		}

		// 章节 XHTML
		chapterPath := filepath.Join(outputPath, fmt.Sprintf("OEBPS/Text/chapter-%03v.xhtml", i))
		if err := os.MkdirAll(filepath.Dir(chapterPath), 0755); err != nil {
			return fmt.Errorf("failed to create chapter directory: %v", err)
		}
		file, err := os.Create(chapterPath)
		if err != nil {
			return fmt.Errorf("failed to create chapter file: %v", err)
		}
		defer file.Close()

		// 修正 HTML 里图片相对路径
		text := chapter.Content.Html
		for _, imgName := range imageNames {
			text = strings.ReplaceAll(text, imgName, fmt.Sprintf("../Images/chapter-%03v/%s", i, imgName))
		}
		if err := template.ContentXHTML(chapter.Title, text).Render(context.Background(), file); err != nil {
			return fmt.Errorf("failed to write chapter: %v", err)
		}
	}

	// 将 Cover 写入（若上游没提供，Cover/Url 可能为空，此时仍会生成 cover.<ext>；你可按需加判空）
	coverExt := strings.TrimPrefix(filepath.Ext(volume.CoverUrl), ".")
	if coverExt == "" {
		coverExt = "jpg"
	}
	coverPath := filepath.Join(outputPath, fmt.Sprintf("cover.%s", strings.ReplaceAll(coverExt, "jpeg", "jpg")))
	if err := os.WriteFile(coverPath, volume.Cover, 0644); err != nil {
		return fmt.Errorf("failed to write cover: %v", err)
	}

	// 写 CoverXHTML 到 OEBPS/Text/cover.xhtml
	coverXHTMLPath := filepath.Join(outputPath, "OEBPS/Text/cover.xhtml")
	file, err := os.Create(coverXHTMLPath)
	if err != nil {
		return fmt.Errorf("failed to create cover XHTML file: %v", err)
	}
	defer file.Close()
	if err := template.CoverXHTML(fmt.Sprintf("../../%s", filepath.Base(coverPath))).Render(context.Background(), file); err != nil {
		return fmt.Errorf("failed to render cover XHTML: %v", err)
	}

	// 目录页 OEBPS/Text/contents.xhtml
	contentsXHTMLPath := filepath.Join(outputPath, "OEBPS/Text/contents.xhtml")
	file, err = os.Create(contentsXHTMLPath)
	if err != nil {
		return fmt.Errorf("failed to create contents XHTML file: %v", err)
	}
	defer file.Close()
	var contents strings.Builder
	contents.WriteString(`<nav epub:type="toc" id="toc">`)
	contents.WriteString(`<ol>`)
	for i, chapter := range volume.Chapters {
		if chapter == nil {
			continue
		}
		contents.WriteString(fmt.Sprintf(`<li><a href="chapter-%03v.xhtml">%s</a></li>`, i, chapter.Title))
	}
	contents.WriteString(`</ol>`)
	contents.WriteString(`</nav>`)
	if err := template.ContentXHTML("目录", contents.String()).Render(context.Background(), file); err != nil {
		return fmt.Errorf("failed to render contents XHTML: %v", err)
	}

	// META-INF/container.xml
	containerPath := filepath.Join(outputPath, "META-INF/container.xml")
	if err := os.MkdirAll(filepath.Dir(containerPath), 0755); err != nil {
		return fmt.Errorf("failed to create container directory: %v", err)
	}
	file, err = os.Create(containerPath)
	if err != nil {
		return fmt.Errorf("failed to create container file: %v", err)
	}
	defer file.Close()
	if err := template.ContainerXML().Render(context.Background(), file); err != nil {
		return fmt.Errorf("failed to render container: %v", err)
	}

	// content.opf
	u := uuid.New()
	if err := CreateContentOPF(outputPath, u.String(), volume, extraFiles); err != nil {
		return fmt.Errorf("failed to create content OPF: %v", err)
	}

	// 写入 CSS
	cssPath := filepath.Join(outputPath, "style.css")
	if err := os.WriteFile(cssPath, []byte(styleCSS), 0644); err != nil {
		return fmt.Errorf("failed to write CSS: %v", err)
	}

	// 写入 extraFiles
	for _, ef := range extraFiles {
		extraFilePath := filepath.Join(outputPath, ef.Path)
		if err := os.WriteFile(extraFilePath, ef.Data, 0644); err != nil {
			return fmt.Errorf("failed to write extra file: %v", err)
		}
	}

	// 打包成 .epub
	if err := PackEpub(outputPath); err != nil {
		return fmt.Errorf("failed to pack epub: %v", err)
	}
	return nil
}

// 封面选择：先插图章(HTML顺序第一张)，否则第一个有图的章节(HTML顺序第一张)，兜底为该章Images文件名单序第一张
func chooseCover(volume *model.Volume) {
	reImg := regexp.MustCompile(`(?is)<img[^>]+src=['"]([^'"]+)['"][^>]*>`)
	isIllustration := func(title string) bool {
		return strings.Contains(title, "插图") ||
			strings.Contains(title, "插畫") ||
			strings.Contains(title, "插画") ||
			strings.Contains(title, "口絵") ||
			strings.Contains(title, "口绘")
	}

	// 在一章里按 HTML 顺序匹配第一张
	pickFromHTML := func(ch *model.Chapter) bool {
		if ch == nil {
			return false
		}
		matches := reImg.FindAllStringSubmatch(ch.Content.Html, -1)
		if len(matches) == 0 {
			return false
		}
		for _, mm := range matches {
			if len(mm) < 2 {
				continue
			}
			src := mm[1]
			base := filepath.Base(src)
			for k, data := range ch.Content.Images {
				if filepath.Base(k) == base && len(data) > 0 {
					volume.Cover = data
					volume.CoverUrl = src // 用于推断扩展名
					return true
				}
			}
		}
		return false
	}

	// 在一章里按文件名排序选第一张
	pickFromImagesByName := func(ch *model.Chapter) bool {
		if ch == nil || len(ch.Content.Images) == 0 {
			return false
		}
		keys := make([]string, 0, len(ch.Content.Images))
		for k := range ch.Content.Images {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if data := ch.Content.Images[keys[0]]; len(data) > 0 {
			volume.Cover = data
			volume.CoverUrl = keys[0]
			return true
		}
		return false
	}

	// 1) 插图章节：HTML 顺序第一张 → 对不上则该章文件名第一张
	for i := range volume.Chapters {
		ch := volume.Chapters[i]
		if ch == nil || !isIllustration(ch.Title) {
			continue
		}
		if pickFromHTML(ch) || pickFromImagesByName(ch) {
			return
		}
	}

	// 2) 兜底：卷内顺序第一个“有图”的章节：HTML 顺序第一张 → 对不上则该章文件名第一张
	for i := range volume.Chapters {
		ch := volume.Chapters[i]
		if ch == nil {
			continue
		}
		hasAny := reImg.MatchString(ch.Content.Html) || len(ch.Content.Images) > 0
		if !hasAny {
			continue
		}
		if pickFromHTML(ch) || pickFromImagesByName(ch) {
			return
		}
	}

	// 3) 最终仍未选到：不强制设置，保留 volume.Cover 现状（若上游已有）
}

func CreateContentOPF(outputPath string, uuid string, volume *model.Volume, extraFiles []model.ExtraFile) error {
	// Dublin Core
	creators := make([]model.DCCreator, 0, len(volume.Authors))
	for _, author := range volume.Authors {
		creators = append(creators, model.DCCreator{Value: author})
	}
	dc := &model.DublinCoreMetadata{
		Titles: []model.DCTitle{{Value: volume.Title}},
		Identifiers: []model.DCIdentifier{{
			Value: fmt.Sprintf("urn:uuid:%s", uuid),
			ID:    "book-id",
		}},
		Languages:    []model.DCLanguage{{Value: "zh-CN"}},
		Descriptions: []model.DCDescription{{Value: volume.Description}},
		Creators:     creators,
		Metas: []model.DublinCoreMeta{
			{Name: "cover", Content: "cover"},
			{Property: "dcterms:modified", Value: time.Now().UTC().Format("2006-01-02T15:04:05Z")},
			{Name: "calibre:series", Content: volume.NovelTitle},
			{Name: "calibre:series_index", Content: strconv.Itoa(volume.SeriesIdx)},
		},
	}

	// Manifest
	manifest := &model.Manifest{Items: make([]model.ManifestItem, 0, 128)}
	manifest.Items = append(manifest.Items,
		model.ManifestItem{ID: "cover.xhtml", Link: "OEBPS/Text/cover.xhtml", Media: "application/xhtml+xml"},
		model.ManifestItem{ID: "contents.xhtml", Link: "OEBPS/Text/contents.xhtml", Media: "application/xhtml+xml", Properties: "nav"},
	)

	coverExt := strings.TrimPrefix(filepath.Ext(volume.CoverUrl), ".")
	if coverExt == "" {
		coverExt = "jpg"
	}
	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID:         "cover",
		Link:       fmt.Sprintf("cover.%s", strings.ReplaceAll(coverExt, "jpeg", "jpg")),
		Media:      fmt.Sprintf("image/%s", strings.ReplaceAll(coverExt, "jpg", "jpeg")),
		Properties: "cover-image",
	})

	for i, chapter := range volume.Chapters {
		if chapter == nil {
			continue
		}
		manifest.Items = append(manifest.Items, model.ManifestItem{
			ID:    fmt.Sprintf("chapter-%03v.xhtml", i),
			Link:  fmt.Sprintf("OEBPS/Text/chapter-%03v.xhtml", i),
			Media: "application/xhtml+xml",
		})
		for filename := range chapter.Content.Images {
			item := model.ManifestItem{
				ID:    fmt.Sprintf("chapter-%03v-%s", i, filepath.Base(filename)),
				Link:  fmt.Sprintf("OEBPS/Images/chapter-%03v/%s", i, filepath.Base(filename)),
				Media: fmt.Sprintf("image/%s", strings.ReplaceAll(strings.TrimPrefix(filepath.Ext(filename), "."), "jpg", "jpeg")),
			}
			manifest.Items = append(manifest.Items, item)
		}
	}

	manifest.Items = append(manifest.Items, model.ManifestItem{
		ID: "style", Link: "style.css", Media: "text/css",
	})
	for _, f := range extraFiles {
		manifest.Items = append(manifest.Items, f.ManifestItem)
	}

	// Spine
	spine := &model.Spine{Items: make([]model.SpineItem, 0, len(manifest.Items))}
	for _, item := range manifest.Items {
		if filepath.Ext(item.Link) == ".xhtml" {
			spine.Items = append(spine.Items, model.SpineItem{IDref: item.ID})
		}
	}

	// 写 content.opf
	contentOPFPath := filepath.Join(outputPath, "content.opf")
	if err := os.MkdirAll(path.Dir(contentOPFPath), 0755); err != nil {
		return fmt.Errorf("failed to create content directory: %v", err)
	}
	file, err := os.Create(contentOPFPath)
	if err != nil {
		return fmt.Errorf("failed to create content file: %v", err)
	}
	defer file.Close()
	if err := template.ContentOPF("book-id", dc, manifest, spine, nil).Render(context.Background(), file); err != nil {
		return fmt.Errorf("failed to render content: %v", err)
	}
	return nil
}

func PackEpub(dirPath string) error {
	savePath := strings.TrimSuffix(dirPath, string(filepath.Separator)) + ".epub"
	zipFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// mimetype 必须是 zip 内的第一个条目且不压缩（EPUB 约定）
	if err := addStringToZip(zipWriter, "mimetype", "application/epub+zip", zip.Store); err != nil {
		return err
	}
	// 其余文件正常写入（Deflate）
	if err := addDirContentToZip(zipWriter, dirPath, zip.Deflate); err != nil {
		return err
	}
	return nil
}

func addStringToZip(zipWriter *zip.Writer, relPath, content string, method uint16) error {
	header := &zip.FileHeader{
		Name:   relPath,
		Method: method,
	}
	w, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

// 将目录下所有文件写入 zip，保证条目名使用正斜杠 `/`
func addDirContentToZip(zipWriter *zip.Writer, dirPath string, method uint16) error {
	return filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if filepath.Base(filePath) == "volume.json" {
			return nil
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 计算相对路径并统一为 `/`
		relPath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = method

		w, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		return err
	})
}
