package mf

import (
	"fmt"
	"strings"
	"time"

	"brainbaking.com/go-jamming/common"
	"willnorris.com/go/microformats"
)

const (
	DateFormatWithTimeZone               = "2006-01-02T15:04:05-07:00"
	dateFormatWithAbsoluteTimeZone       = "2006-01-02T15:04:05-0700"
	dateFormatWithTimeZoneSuffixed       = "2006-01-02T15:04:05.000Z"
	dateFormatWithoutTimeZone            = "2006-01-02T15:04:05"
	dateFormatWithSecondsWithoutTimeZone = "2006-01-02T15:04:05.00Z"
	dateFormatWithoutTime                = "2006-01-02"
	Anonymous                            = "anonymous"
)

var (
	// This is similar to Hugo's string-to-date casting system
	// See https://github.com/spf13/cast/blob/master/caste.go
	supportedFormats = []string{
		DateFormatWithTimeZone,
		dateFormatWithAbsoluteTimeZone,
		dateFormatWithTimeZoneSuffixed,
		dateFormatWithSecondsWithoutTimeZone,
		dateFormatWithoutTimeZone,
		dateFormatWithoutTime,
	}
)

type IndiewebAuthor struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (ia *IndiewebAuthor) AnonymizePicture() {
	ia.Picture = fmt.Sprintf("/pictures/%s", Anonymous)
}

func (ia *IndiewebAuthor) AnonymizeName() {
	ia.Name = "Anonymous"
}

type IndiewebDataResult struct {
	Status string          `json:"status"`
	Data   []*IndiewebData `json:"json"`
}

func ResultFailure(data []*IndiewebData) IndiewebDataResult {
	return emptyNilData(IndiewebDataResult{
		Status: "failure",
		Data:   data,
	})
}

func ResultSuccess(data []*IndiewebData) IndiewebDataResult {
	return emptyNilData(IndiewebDataResult{
		Status: "success",
		Data:   data,
	})
}

func emptyNilData(result IndiewebDataResult) IndiewebDataResult {
	if result.Data == nil {
		result.Data = make([]*IndiewebData, 0)
	}
	return result
}

type IndiewebData struct {
	Author       IndiewebAuthor `json:"author"`
	Name         string         `json:"name"`
	Content      string         `json:"content"`
	Published    string         `json:"published"`
	Url          string         `json:"url"`
	IndiewebType MfType         `json:"type"`
	Source       string         `json:"source"`
	Target       string         `json:"target"`
}

func (id *IndiewebData) PublishedDate() time.Time {
	return common.ToTime(id.Published, DateFormatWithTimeZone)
}

func (id *IndiewebData) AsMention() Mention {
	return Mention{
		Source: id.Source,
		Target: id.Target,
	}
}

func (id *IndiewebData) IsEmpty() bool {
	return id.Url == ""
}

func PublishedNow() string {
	return common.Now().UTC().Format(DateFormatWithTimeZone)
}

// Go stuff: entry.Properties["name"][0].(string),
// JS stuff: hEntry.properties?.name?.[0]
// The problem: convoluted syntax and no optional chaining!
func Str(mf *microformats.Microformat, key string) string {
	val := mf.Properties[key]
	if len(val) == 0 {
		return ""
	}

	str, ok := val[0].(string)
	if !ok {
		// in very weird cases, it could be a map holding a value, like in mf2's "photo"
		valMap, ok2 := val[0].(map[string]string)
		if !ok2 {
			return ""
		}
		return valMap["value"]
	}

	return str
}

func Map(mf *microformats.Microformat, key string) map[string]string {
	val := mf.Properties[key]
	if len(val) == 0 {
		return map[string]string{}
	}
	mapVal, ok := val[0].(map[string]string)
	if !ok {
		return map[string]string{}
	}
	return mapVal
}

func HEntry(data *microformats.Data) *microformats.Microformat {
	return hItemType(data, "h-entry")
}

func HCard(data *microformats.Data) *microformats.Microformat {
	return hItemType(data, "h-card")
}

func hItemType(data *microformats.Data, hType string) *microformats.Microformat {
	for _, itm := range data.Items {
		if common.Includes(itm.Type, hType) {
			return itm
		}
	}
	return nil
}

func mfEmpty() *microformats.Microformat {
	return &microformats.Microformat{
		Properties: map[string][]interface{}{},
	}
}

func Prop(mf *microformats.Microformat, key string) *microformats.Microformat {
	val := mf.Properties[key]
	if len(val) == 0 {
		return mfEmpty()
	}
	for i := range val {
		conv, ok := val[i].(*microformats.Microformat)
		if ok {
			return conv
		}
	}
	return mfEmpty()
}

func Published(hEntry *microformats.Microformat) string {
	publishedDate := Str(hEntry, "published")
	if publishedDate == "" {
		return PublishedNow()
	}

	for _, format := range supportedFormats {
		formatted, err := time.Parse(format, publishedDate)
		if err != nil {
			continue
		}
		return formatted.Format(DateFormatWithTimeZone)
	}

	return PublishedNow()
}

func NewAuthor(hEntry *microformats.Microformat, hCard *microformats.Microformat) IndiewebAuthor {
	name := ""
	if hCard != nil {
		name = DetermineAuthorName(hCard)
	}
	if name == "" {
		name = DetermineAuthorName(hEntry)
	}
	picture := DetermineAuthorPhoto(hEntry)
	if picture == "" {
		picture = DetermineAuthorPhoto(hCard)
	}
	return IndiewebAuthor{
		Picture: picture,
		Name:    name,
	}
}

func DetermineAuthorPhoto(hEntry *microformats.Microformat) string {
	photo := Str(Prop(hEntry, "author"), "photo")
	if photo == "" {
		photo = Str(hEntry, "photo")
	}
	return photo
}

func DetermineAuthorName(hEntry *microformats.Microformat) string {
	authorName := Str(Prop(hEntry, "author"), "name")
	if authorName == "" {
		authorName = Prop(hEntry, "author").Value
	}
	if authorName == "" {
		authorName = Str(hEntry, "author")
	}
	if authorName == "" {
		authorName = Str(hEntry, "name")
	}
	return authorName
}

type MfType string

const (
	TypeLink     MfType = "link"
	TypeReply    MfType = "reply"
	TypeRepost   MfType = "repost"
	TypeLike     MfType = "like"
	TypeBookmark MfType = "bookmark"
	TypeMention  MfType = "mention"
)

func Type(hEntry *microformats.Microformat) MfType {
	hType := Str(hEntry, "like-of")
	if hType != "" {
		return TypeLike
	}
	hType = Str(hEntry, "bookmark-of")
	if hType != "" {
		return TypeBookmark
	}
	hType = Str(hEntry, "repost-of")
	if hType != "" {
		return TypeRepost
	}
	hType = Str(hEntry, "in-reply-to")
	if hType != "" {
		return TypeReply
	}
	return TypeMention
}

// Mastodon uids start with "tag:server", but we do want indieweb uids from other sources
func Url(hEntry *microformats.Microformat, source string) string {
	uid := Str(hEntry, "uid")
	if uid != "" && strings.HasPrefix(uid, "http") {
		return uid
	}
	url := Str(hEntry, "url")
	if url != "" {
		return url
	}
	return source
}

func Content(hEntry *microformats.Microformat) string {
	bridgyTwitterContent := Str(hEntry, "bridgy-twitter-content")
	if bridgyTwitterContent != "" {
		return common.Shorten(bridgyTwitterContent)
	}
	summary := Str(hEntry, "summary")
	if summary != "" {
		return common.Shorten(summary)
	}
	contentEntry := Map(hEntry, "content")["value"]
	return common.Shorten(contentEntry)
}
