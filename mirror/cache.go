package mirror

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	cacheExpiration = 12 * time.Hour
)

//go:embed index.html
var home string

type Cache struct {
	settings  settings
	createdAt time.Time
	template  *template.Template
	HTML      []byte
	JSON      []byte
}

func (c *Cache) isExpired() bool {
	return time.Since(c.createdAt) > cacheExpiration
}

type JSONResponse struct {
	Data []Group `json:"data"`
}

func (c *Cache) refresh() error {
	var fs []File
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(c.settings.region),
		Endpoint:         aws.String(c.settings.endpointURL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			c.settings.accessKey,
			c.settings.secretAccessKey,
			"",
		),
	})
	if err != nil {
		return err
	}

	var token *string
	loadPage := func(t *string) ([]File, *string, error) {
		var fs []File
		sdk := s3.New(sess)
		r, err := sdk.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(c.settings.bucket),
			ContinuationToken: t,
		})
		if err != nil {
			return []File{}, nil, err
		}
		for _, obj := range r.Contents {
			url := fmt.Sprintf("%s%s", c.settings.publicDomain, *obj.Key)
			fs = append(fs, File{url, *obj.Size, *obj.Key, *obj.LastModified})
		}
		if *r.IsTruncated {
			return fs, r.NextContinuationToken, nil
		}
		return fs, nil, nil
	}
	for {
		r, nxt, err := loadPage(token)
		if err != nil {
			return err
		}
		fs = append(fs, r...)
		if nxt == nil {
			break
		}
		token = nxt
	}

	data := newGroups(fs)
	var h bytes.Buffer
	c.template.Execute(&h, data)
	c.HTML = h.Bytes()

	var j bytes.Buffer
	if err := json.NewEncoder(&j).Encode(JSONResponse{data}); err != nil {
		return err
	}
	c.JSON = j.Bytes()

	c.createdAt = time.Now()
	return nil
}

func newCache(s settings) (*Cache, error) {
	t, err := template.New("home").Parse(home)
	if err != nil {
		return nil, err
	}
	c := Cache{s, time.Now(), t, []byte{}, []byte{}}
	if err := c.refresh(); err != nil {
		return nil, err
	}
	return &c, nil
}
