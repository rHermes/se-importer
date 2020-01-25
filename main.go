package main

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/lzmadec"

	"github.com/gen2brain/go-unarr"

	"github.com/alecthomas/repr"

	mssql "github.com/denisenkom/go-mssqldb"

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

// Use to execute function on one file in the 7zip archive.
func execOn7zFile(fpath, name string, f func(io.Reader) error) error {
	sz, err := go7z.OpenReader(fpath)
	if err != nil {
		return err
	}
	defer sz.Close()

	for {
		hdr, err := sz.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if hdr.Name != name {
			if _, err := io.Copy(ioutil.Discard, sz); err != nil {
				return err
			}
			continue
		}
		// We should get a file
		if hdr.IsEmptyStream {
			return errors.New("This should not be empty!")
		}

		if err := f(sz); err != nil {
			return err
		}
		return nil
	}

	return errors.New("File not found!")
}

// Use to execute function on one file in the 7zip archive.
func execOn7zFilePro(fpath, name string, f func(io.Reader) error) error {
	a, err := unarr.NewArchive(fpath)
	if err != nil {
		return err
	}
	defer a.Close()

	list, err := a.List()
	if err != nil {
		return err
	}
	repr.Println(list)

	if err := a.EntryFor(name); err != nil {
		return err
	}

	if err := f(a); err != nil {
		return err
	}

	return nil
}

// Use to execute function on one file in the 7zip archive.
func execOn7zFileProer(fpath, name string, f func(io.Reader) error) error {
	a, err := lzmadec.NewArchive(fpath)
	if err != nil {
		return err
	}

	rdr, err := a.GetFileReader(name)
	if err != nil {
		return err
	}
	defer rdr.Close()

	if err := f(rdr); err != nil {
		return err
	}

	return nil
}

func ParseStack7zSQL(db *sql.DB, name, fpath string) error {
	tt := time.Now()
	fmt.Printf("%-40s", name)

	// Now we setup the sites
	if _, err := db.Exec(sqlSetupTables); err != nil {
		repr.Println(err)
		return err
	}

	// Now to drop the data we have
	dSite, err := db.Prepare(sqlDeleteSite)
	if err != nil {
		return err
	}
	defer dSite.Close()
	if _, err := dSite.Exec(sql.Named("name", name)); err != nil {
		return err
	}

	// First we need to insert the site
	iSite, err := db.Prepare(sqlInsertSite)
	if err != nil {
		return err
	}
	defer iSite.Close()

	var siteID int
	row := iSite.QueryRow(sql.Named("name", name))
	if err := row.Scan(&siteID); err != nil {
		repr.Println(err)
		return err
	}

	//log.Printf("The site id for %s is %d\n", name, siteID)

	// This is simpler.
	fUser := func(r io.Reader) error {
		//log.Printf("Beginning to parse users\n")
		decoder := xml.NewDecoder(r)

		txn, err := db.Begin()
		if err != nil {
			return err
		}
		defer txn.Rollback()

		// We create the insert statement
		stmt, err := txn.Prepare(mssql.CopyIn(
			"[user]",
			mssql.BulkOptions{},
			"id",
			"site_id",
			"reputation",
			"creation_date",
			"display_name",
			"last_access_date",
			"website_url",
			"location",
			"about_me",
			"views",
			"up_votes",
			"down_votes",
			"profile_image_url",
			"account_id",
		))
		if err != nil {
			return err
		}
		defer stmt.Close()

		inusers := false
	L:
		for {
			t, err := decoder.Token()
			if err != nil {
				return err
			}

			switch se := t.(type) {
			case xml.StartElement:
				if !inusers {
					if se.Name.Local != "users" {
						return errors.New("this aint goot")
					}
					inusers = true
					break
				}

				if se.Name.Local != "row" {
					return errors.New("Invalid element name: " + se.Name.Local)
				}

				var user User
				if err := decoder.DecodeElement(&user, &se); err != nil {
					return err
				}

				_, err = stmt.Exec(
					user.ID,
					siteID,
					user.Reputation,
					user.CreationDate.Time,
					user.DisplayName,
					user.LastAccessDate.Time,
					user.WebsiteURL,
					user.Location,
					user.AboutMe,
					user.Views,
					user.UpVotes,
					user.DownVotes,
					user.ProfileImageURL,
					user.AccountID,
				)
				if err != nil {
					return err
				}

			case xml.EndElement:
				if inusers {
					if se.Name.Local == "users" {
						inusers = false
						break L
					}
				}
			}
		}

		res, err := stmt.Exec()
		if err != nil {
			return err
		}

		if err := txn.Commit(); err != nil {
			return err
		}

		rowCount, err := res.RowsAffected()
		if err != nil {
			return err
		}
		rowCount = rowCount
		//log.Printf("We copied %d rows.\n", rowCount)
		//log.Printf("Done parsing users\n")
		return nil
	}

	if err := execOn7zFileProer(fpath, "Users.xml", fUser); err != nil {
		return err
	}

	dd := time.Since(tt)
	fmt.Printf(" %10d ms\n", dd.Milliseconds())
	return nil
}

func makeConnURL() *url.URL {
	v := url.Values{}
	v.Set("database", os.Getenv("MSSQL_DB"))
	return &url.URL{
		Scheme:   "sqlserver",
		Host:     os.Getenv("MSSQL_HOST"),
		User:     url.UserPassword(os.Getenv("MSSQL_USER"), os.Getenv("MSSQL_PASSWD")),
		RawQuery: v.Encode(),
	}
}

func main() {
	// processSingleArchive()
	if err := processWholeFolder(); err != nil {
		log.Fatalf("Couldn't process folder: %s\n", err.Error())
	}
}

func processWholeFolder() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("need to provide directory to iterate")
	}


	connStr := makeConnURL().String()
	connector, err := mssql.NewConnector(connStr)
	if err != nil {
		return fmt.Errorf("Error creating connector: %s", err.Error())
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	// We try a ping here, just to see
	if err := db.Ping(); err != nil {
		return fmt.Errorf("We could not ping the database: %s", err.Error())
	}

	ents, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		return err
	}

	for _, ent := range ents {
		nm := ent.Name()
		if strings.HasSuffix(nm, ".stackexchange.com.7z") && !ent.IsDir() {
			name := strings.TrimSuffix(nm, ".stackexchange.com.7z")
			fpath := filepath.Join(os.Args[1], nm)
			if err := ParseStack7zSQL(db, name, fpath); err != nil {
				return err
			}
		}
	}
	return nil
}

func processSingleArchive() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("need to provide name then path")

	}

	connStr := makeConnURL().String()
	connector, err := mssql.NewConnector(connStr)
	if err != nil {
		return fmt.Errorf("Error creating connector: %s", err.Error())
	}

	db := sql.OpenDB(connector)
	defer db.Close()

	// We try a ping here, just to see
	if err := db.Ping(); err != nil {
		return fmt.Errorf("We could not ping the database: %s", err.Error())
	}

	// open up 7zip file
	if err := ParseStack7zSQL(db, os.Args[1], os.Args[2]); err != nil {
		return fmt.Errorf("Couldn't parse 7z: %s", err.Error())
	}

	return nil
}
