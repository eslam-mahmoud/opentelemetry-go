package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"go.opentelemetry.io/otel/sdk/export/trace"
)

// Exporter is a trace exporter that uploads data.
type Exporter struct {
	backoff           int // number of seconds between requests
	lastCommunication int
	endpoint          string
	apiKey            string
	serviceName       string
	debugging         bool
	spans             []*trace.SpanData
	logger            kitlog.Logger
}

// Options constructor options for Exporter
type Options struct {
	Endpoint    string
	APIKey      string
	ServiceName string
	Debugging   bool
	Logger      kitlog.Logger
}

// NewExporter constructor for Exporter
func NewExporter(opts Options) (*Exporter, error) {
	if opts.Logger == nil {
		opts.Logger = kitlog.With(kitlog.NewJSONLogger(os.Stderr), "ts", kitlog.DefaultTimestampUTC)
	}

	return &Exporter{
		endpoint:    opts.Endpoint,
		apiKey:      opts.APIKey,
		serviceName: opts.ServiceName,
		debugging:   opts.Debugging,
		logger:      opts.Logger,
	}, nil
}

// ExportSpan exports a SpanData.
func (e *Exporter) ExportSpan(ctx context.Context, span *trace.SpanData) {
	// always send array because the API endpoint can take more than one span
	// and to be able to cache spans and send them as patchs

	// TODO cache by parent id & name with count of requests for whenever going to emplement sampling on the client
	e.spans = append(e.spans, span)
	e.sendSpans(ctx)
}

// ExportSpans getting many spans to be exported
func (e *Exporter) ExportSpans(ctx context.Context, spans []*trace.SpanData) {
	// TODO cache by parent id & name with count of requests for whenever going to emplement sampling on the client
	e.spans = append(e.spans, spans...)
	e.sendSpans(ctx)
}

func (e *Exporter) sendSpans(ctx context.Context) {
	// TODO check when was the last time we did send request
	// if time is too small cache span and send them as patch

	// encode
	spanBytes, err := json.Marshal(e.spans)
	if err != nil {
		e.logger.Log(
			"message", "could not Marshal spans body",
			"severity", "CRITICAL",
			"spans", e.spans,
			"err", err,
		)
		return
	}
	if e.debugging {
		e.logger.Log(
			"message", "debugging spans body",
			"severity", "DEBUG",
			"spans", e.spans,
			"backoff", e.backoff,
		)
	}

	// send HTTP req to the endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint, bytes.NewBuffer(spanBytes))
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if e.backoff == 0 {
			e.backoff = 1
		} else {
			e.backoff = 2 * e.backoff
		}
		e.logger.Log(
			"message", "could not send spans body to API",
			"severity", "CRITICAL",
			"spans", e.spans,
			"err", err,
			"backoff", e.backoff,
		)
		// spans are cached and will be sent with the next req
		return
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body) // TODO defer ?!
	// TODO update latest communication field to use backoff and caching

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
}

// Shutdown TODO empliment
func (e *Exporter) Shutdown() {

}
