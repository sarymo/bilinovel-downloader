package model

import "encoding/xml"

type TocNCXHead struct {
	XMLName xml.Name         `xml:"head"`
	Meta    []TocNCXHeadMeta `xml:"meta"`
}

type TocNCXHeadMeta struct {
	XMLName xml.Name `xml:"meta"`
	Content string   `xml:"content,attr"`
	Name    string   `xml:"name,attr"`
}

func (h *TocNCXHead) Marshal() (string, error) {
	xmlBytes, err := xml.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(xmlBytes), nil
}

type NavPoint struct {
	Id        string          `xml:"id,attr"`
	PlayOrder int             `xml:"playOrder,attr"`
	Label     string          `xml:"navLabel>text"`
	Content   NavPointContent `xml:"content"`
	NavPoints []*NavPoint     `xml:"navPoint"`
}

type NavPointContent struct {
	Src string `xml:"src,attr"`
}

type NavMap struct {
	XMLName xml.Name    `xml:"navMap"`
	Points  []*NavPoint `xml:"navPoint"`
}

func (n *NavMap) Marshal() (string, error) {
	xmlBytes, err := xml.Marshal(n)
	if err != nil {
		return "", err
	}
	return string(xmlBytes), nil
}
