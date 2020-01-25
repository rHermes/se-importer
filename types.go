package main

import (
	"encoding/xml"
	"time"
)

// Time does not implment unmarshal, so I had to do this
type SEDate struct {
	time.Time
}

func (t *SEDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	ti, err := time.Parse("2006-01-02T15:04:05.999", s)
	if err != nil {
		return err
	}
	t.Time = ti
	return nil
}

func (t *SEDate) UnmarshalXMLAttr(attr xml.Attr) error {
	ti, err := time.Parse("2006-01-02T15:04:05.999", attr.Value)
	if err != nil {
		return err
	}
	t.Time = ti
	return nil
}

type Site struct {
	Users    []User
	Tags     []Tag
	Badges   []Badge
	Posts    []Post
	Comments []Comment
}

type User struct {
	ID              int     `xml:"Id,attr"`
	Reputation      int     `xml:"Reputation,attr"`
	CreationDate    SEDate  `xml:"CreationDate,attr"`
	DisplayName     *string `xml:"DisplayName,attr"`
	LastAccessDate  SEDate  `xml:"LastAccessDate,attr"`
	WebsiteURL      *string `xml:"WebsiteUrl,attr"`
	Location        *string `xml:"Location,attr"`
	AboutMe         *string `xml:"AboutMe,attr"`
	Views           int     `xml:"Views,attr"`
	UpVotes         int     `xml:"UpVotes,attr"`
	DownVotes       int     `xml:"DownVotes,attr"`
	ProfileImageURL *string `xml:"ProfileImageUrl,attr"`
	AccountID       *int    `xml:"AccountId,attr"`
}

type Tag struct {
	ID            int    `xml:"Id,attr"`
	TagName       string `xml:"TagName,attr"`
	Count         int    `xml:"Count,attr"`
	ExcerptPostID *int   `xml:"ExcerptPostId,attr"`
	WikiPostID    *int   `xml:"WikiPostId,attr"`
}

type Badge struct {
	ID       int    `xml:"Id,attr"`
	UserID   int    `xml:"UserId,attr"`
	Name     string `xml:"Name,attr"`
	Date     SEDate `xml:"Date,attr"`
	Class    int    `xml:"Class,attr"`
	TagBased bool   `xml:"TagBased,attr"`
}

type Post struct {
	ID                 int    `xml:"Id,attr"`
	PostTypeID         int    `xml:"PostTypeId,attr"`
	AcceptedAnswerID   *int   `xml:"AcceptedAnswerId,attr"`
	ParentID           *int   `xml:"ParentId,attr"`
	CreationDate       SEDate `xml:"CreationDate,attr"`
	Score              int    `xml:"Score,attr"`
	ViewCount          *int   `xml:"ViewCount,attr"`
	Body               string `xml:"Body,attr"`
	OwnerUserID        *int   `xml:"OwnerUserId,attr"`
	LastActivityDate   SEDate `xml:"LastActivityDate,attr"`
	Title              string `xml:"Title,attr"`
	Tags               string `xml:"Tags,attr"`
	AnswerCount        string `xml:"AnswerCount,attr"`
	CommentCount       string `xml:"CommentCount,attr"`
	FavoriteCount      string `xml:"FavoriteCount,attr"`
	LastEditorUserId   string `xml:"LastEditorUserId,attr"`
	LastEditDate       SEDate `xml:"LastEditDate,attr"`
	CommunityOwnedDate SEDate `xml:"CommunityOwnedDate,attr"`
	ClosedDate         SEDate `xml:"ClosedDate,attr"`
	OwnerDisplayName   string `xml:"OwnerDisplayName,attr"`
}

type Comment struct {
	ID              int     `xml:"Id,attr"`
	PostID          int     `xml:"PostId,attr"`
	Score           int     `xml:"Score,attr"`
	Text            string  `xml:"Text,attr"`
	CreationDate    SEDate  `xml:"CreationDate,attr"`
	UserDisplayName *string `xml:"UserDisplayName,attr"`
	UserID          *int    `xml:"UserId,attr"`
}
