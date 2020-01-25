package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	mssql "github.com/denisenkom/go-mssqldb"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/repr"

	"github.com/saracen/go7z"
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
	Users  []User
	Tags   []Tag
	Badges []Badge
	Posts  []Post
	Comments []Comment
}

type User struct {
	ID              int     `xml:"Id,attr"`
	Reputation      int     `xml:"Reputation,attr"`
	CreationDate    SEDate  `xml:"CreationDate,attr"`
	DisplayName     string  `xml:"DisplayName,attr"`
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
	UserID          *int     `xml:"UserId,attr"`
}

func ParseUsers(r io.Reader) ([]User, error) {
	xd := xml.NewDecoder(r)

	t := time.Now()
	var ud struct {
		XMLName xml.Name `xml:"users"`
		Users   []User   `xml:"row"`
	}
	if err := xd.Decode(&ud); err != nil {
		return nil, err
	}
	log.Printf("PERF: Parsing users took %s\n", time.Since(t).String())

	return ud.Users, nil
}

func ParseTags(r io.Reader) ([]Tag, error) {
	xd := xml.NewDecoder(r)
	t := time.Now()
	var ud struct {
		XMLName xml.Name `xml:"tags"`
		Tags    []Tag    `xml:"row"`
	}
	if err := xd.Decode(&ud); err != nil {
		return nil, err
	}
	log.Printf("PERF: Parsing tags took %s\n", time.Since(t).String())

	return ud.Tags, nil
}

func ParseBadges(r io.Reader) ([]Badge, error) {
	xd := xml.NewDecoder(r)
	t := time.Now()
	var ud struct {
		XMLName xml.Name `xml:"badges"`
		Badges  []Badge  `xml:"row"`
	}
	if err := xd.Decode(&ud); err != nil {
		return nil, err
	}
	log.Printf("PERF: Parsing badges took %s\n", time.Since(t).String())

	return ud.Badges, nil
}

func ParsePosts(r io.Reader) ([]Post, error) {
	xd := xml.NewDecoder(r)
	t := time.Now()
	var ud struct {
		XMLName xml.Name `xml:"posts"`
		Posts   []Post   `xml:"row"`
	}
	if err := xd.Decode(&ud); err != nil {
		return nil, err
	}
	log.Printf("PERF: Parsing posts took %s\n", time.Since(t).String())

	return ud.Posts, nil
}

func ParseComments(r io.Reader) ([]Comment, error) {
	xd := xml.NewDecoder(r)
	t := time.Now()
	var ud struct {
		XMLName xml.Name `xml:"comments"`
		Comments   []Comment  `xml:"row"`
	}
	if err := xd.Decode(&ud); err != nil {
		return nil, err
	}
	log.Printf("PERF: Parsing comments took %s\n", time.Since(t).String())

	return ud.Comments, nil
}

func ParseStack7z(fpath string) (*Site, error) {
	sz, err := go7z.OpenReader(fpath)
	if err != nil {
		return nil, err
	}
	defer sz.Close()

	var s Site
	for {
		hdr, err := sz.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// We should get a file
		if hdr.IsEmptyStream {
			log.Printf("WARN: We should have none of these?\n")
			continue
		}

		log.Printf("Parsing entry: %s\n", hdr.Name)

		switch hdr.Name {
		case "Users.xml":
			s.Users, err = ParseUsers(sz)
		case "Tags.xml":
			s.Tags, err = ParseTags(sz)
		case "Badges.xml":
			s.Badges, err = ParseBadges(sz)
		case "Posts.xml":
			s.Posts, err = ParsePosts(sz)
		case "Comments.xml":
			s.Comments, err = ParseComments(sz)
		default:
			// We need to read the entire file, even if we are just using a few of them
			if _, err := io.Copy(ioutil.Discard, sz); err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return &s, nil
}


func makeConnURL() *url.URL {
	v := url.Values{}
	v.Set("database", os.Getenv("MSSQL_DB"))
	return &url.URL{
		Scheme: "sqlserver",
		Host: os.Getenv("MSSQL_HOST"),
		User: url.UserPassword(os.Getenv("MSSQL_USER"),os.Getenv("MSSQL_PASSWD")),
		RawQuery: v.Encode(),
	}
}


func main() {
	if len(os.Args) != 3 {
		fmt.Printf("need to provide name then path\n")
		return
	}

	connStr := makeConnURL().String()
	connector, err := mssql.NewConnector(connStr)
	if err != nil {
		log.Fatalf("Error creating connector: %s\n", err.Error())
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	// We try a ping here, just to see
	if err := db.Ping(); err != nil {
		log.Fatalf("We could not ping the database: %s\n", err.Error())
	}

	// open up 7zip file
	site, err := ParseStack7z(os.Args[1])
	if err != nil {
		log.Fatalf("Couldn't parse 7z: %s\n", err.Error())
	}

	fmt.Printf("Site info:\n")
	fmt.Printf("  Users: %d\n", len(site.Users))
	fmt.Printf("  Tags: %d\n", len(site.Tags))
	fmt.Printf("  Badges: %d\n", len(site.Badges))
	fmt.Printf("  Posts: %d\n", len(site.Posts))

	for _, post := range site.Posts {
		post = post
		if post.OwnerUserID == nil {
			repr.Println(post)
		}
	}

}
