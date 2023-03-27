package processors

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/derfenix/webarchive/entity"
)

type processor interface {
	Process(ctx context.Context, url string) ([]entity.File, error)
}

func NewProcessors() (*Processors, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Second * 10,
				KeepAlive: time.Second * 10,
			}).DialContext,
			MaxIdleConns:           20,
			MaxIdleConnsPerHost:    5,
			MaxConnsPerHost:        10,
			IdleConnTimeout:        time.Second * 60,
			ResponseHeaderTimeout:  time.Second * 20,
			MaxResponseHeaderBytes: 1024 * 1024 * 50,
			WriteBufferSize:        256,
			ReadBufferSize:         1024 * 64,
			ForceAttemptHTTP2:      true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 3 {
				return fmt.Errorf("too many redirects")
			}

			return nil
		},
		Jar:     jar,
		Timeout: time.Second * 30,
	}

	procs := Processors{
		processors: map[entity.Format]processor{
			entity.FormatHeaders: NewHeaders(httpClient),
			entity.FormatPDF:     NewPDF(),
		},
	}

	return &procs, nil
}

type Processors struct {
	processors map[entity.Format]processor
}

func (p *Processors) Process(ctx context.Context, format entity.Format, url string) entity.Result {
	result := entity.Result{Format: format}

	proc, ok := p.processors[format]
	if !ok {
		result.Err = fmt.Errorf("no processor registered")

		return result
	}

	files, err := proc.Process(ctx, url)
	if err != nil {
		result.Err = fmt.Errorf("process: %w", err)

		return result
	}

	result.Files = files

	return result
}

func (p *Processors) OverrideProcessor(format entity.Format, proc processor) error {
	p.processors[format] = proc

	return nil
}
