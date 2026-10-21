// Package processors archives a page in the supported formats and reads the
// meta of a page. It is the outbound adapter behind application.Archiver.
package processors

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	"go.uber.org/zap"

	"github.com/derfenix/webarchive/internal/config"
	"github.com/derfenix/webarchive/internal/domain"
)

// processor archives one format. Formats share the page content through Source,
// so a page is downloaded once per archive.
type processor interface {
	Process(ctx context.Context, source domain.Source) ([]domain.File, error)
}

func NewProcessors(cfg config.PDF, log *zap.Logger) (*Processors, error) {
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
		client: httpClient,
		log:    log,
		processors: map[domain.Format]processor{
			domain.FormatHeaders:    NewHeaders(httpClient),
			domain.FormatPDF:        NewPDF(cfg),
			domain.FormatSingleFile: NewSingleFile(httpClient, log),
		},
	}

	return &procs, nil
}

type Processors struct {
	processors map[domain.Format]processor
	client     *http.Client
	log        *zap.Logger
}

func (p *Processors) Archive(ctx context.Context, format domain.Format, source domain.Source) ([]domain.File, error) {
	proc, ok := p.processors[format]
	if !ok {
		return nil, fmt.Errorf("no processor registered")
	}

	files, err := proc.Process(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	return files, nil
}
