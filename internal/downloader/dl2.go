package downloader

import (
	pb "github.com/cheggaaa/pb/v3"
	"github.com/hashicorp/go-getter"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"io"
	"net/url"
	"path/filepath"

	"sync"
)

func Download(dlUrl string, targetDir string) error {
	// make sure url is valid
	_, err := url.Parse(dlUrl)
	if err != nil {
		return err
	}
	var notex bool

	if notex, err = IsEmptyDir(targetDir); notex || err != nil {
		httpGetter := getter.HttpGetter{Netrc: true}
		pbar := ProgressBar{}
		progress := []getter.ClientOption{getter.WithProgress(&pbar)}
		if osutil.GetOS() == "windows" {
			progress = []getter.ClientOption{}
		}
		client := &getter.Client{
			Src:     dlUrl,
			Dst:     targetDir,
			Mode:    getter.ClientModeAny,
			Options: progress,
			Getters: map[string]getter.Getter{
				"file":  &getter.FileGetter{Copy: false},
				"http":  &httpGetter,
				"https": &httpGetter,
				"s3":    new(getter.S3Getter),
				"gcs":   new(getter.GCSGetter),
				"git":   new(getter.GitGetter),
			},
		}
		return client.Get()
	}
	return nil
}

type ProgressBar struct {
	lock     sync.Mutex
	progress *pb.ProgressBar
}

// TrackProgress instantiates a new progress bar that will
// display the progress of stream until closed.
// total can be 0.

const pbTpl pb.ProgressBarTemplate = `{{ string . "prefix" }} {{counters . | green }} {{ bar . "<" "=" (cycle . "↖" "↗" "↘" "↙" ) "." ">" | cyan }} {{speed . | green }} {{percent .}} {{ string . "suffix" }}`

func (cpb *ProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()
	if cpb.progress == nil {
		cpb.progress = pb.New64(totalSize)
	}
	p := pbTpl.Start64(totalSize)
	p.Set("prefix", filepath.Base(src+": "))
	p.Set("suffix", "\n")
	p.SetCurrent(currentSize)
	p.Set(pb.Bytes, true)
	barReader := p.NewProxyReader(stream)

	return &readCloser{
		Reader: barReader,
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()
			p.Finish()
			return nil
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (c *readCloser) Close() error { return c.close() }
