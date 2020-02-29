package exporter

import (
	"context"
	"encoding/json"
	"fmt"

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
	// TODO
	// logger            logger
}

// Options constructor options for Exporter
type Options struct {
	Endpoint    string
	APIKey      string
	ServiceName string
	Debugging   bool
	// TODO
	// logger            logger
}

// NewExporter constructor for Exporter
func NewExporter(opts Options) (*Exporter, error) {
	return &Exporter{
		endpoint:    opts.Endpoint,
		apiKey:      opts.APIKey,
		serviceName: opts.ServiceName,
		backoff:     1,
		debugging:   opts.Debugging,
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
		// TODO log
		// fmt.Println(err)
		return
	}
	if e.debugging {
		// TOD should use logger
		fmt.Println(string(spanBytes))
	}

	// // send HTTP req to the endpoint
	// req, err := http.NewRequest("POST", e.endpoint, bytes.NewBuffer(spanBytes))
	// req.Header.Set("Authorization", "Bearer "+e.apiKey)
	// req.Header.Set("Content-Type", "application/json")
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	// TODO log
	// 	// TODO cache the req and attemp resent it bassed on the response status code
	// 	e.backoff = 2 * e.backoff
	// 	return
	// }
	// defer resp.Body.Close()
	// ioutil.ReadAll(resp.Body) // TODO defer ?!
	// // TODO update latest communication field to use backoff and caching

	// // fmt.Println("response Status:", resp.Status)
	// // fmt.Println("response Headers:", resp.Header)
	// // body, _ := ioutil.ReadAll(resp.Body)
	// // fmt.Println("response Body:", string(body))
}

// Shutdown TODO empliment
func (e *Exporter) Shutdown() {

}
