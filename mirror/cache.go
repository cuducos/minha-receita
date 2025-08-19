package mirror

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json/v2"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const cacheExpiration = 12 * time.Hour

//go:embed index.html
var home string

type JSONResponse struct {
	Data []Group `json:"data"`
}

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

func (c *Cache) refresh() error {
	var fs []File
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(c.settings.region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.settings.accessKey,
			c.settings.secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return err
	}

	var t *string
	loadPage := func(t *string) ([]File, *string, error) {
		var fs []File
		sdk := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
			if c.settings.endpointURL != "" {
				o.BaseEndpoint = aws.String(c.settings.endpointURL)
			}
		})
		r, err := sdk.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
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
		r, nxt, err := loadPage(t)
		if err != nil {
			return err
		}
		fs = append(fs, r...)
		if nxt == nil {
			break
		}
		t = nxt
	}

	g := newGroups(fs)
	errs := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		var b bytes.Buffer
		c.template.Execute(&b, g)
		c.HTML = b.Bytes()
	}()
	go func() {
		defer wg.Done()
		c.JSON, err = json.Marshal(JSONResponse{g})
		if err != nil {
			errs <- err
		}
	}()
	go func() {
		wg.Wait()
		close(errs)
	}()
	for err := range errs {
		if err != nil {
			return err
		}
	}
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
