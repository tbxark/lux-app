package luxdownloader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iawia002/lux/extractors"
)

func TestDownloadReportsByteProgress(t *testing.T) {
	body := []byte("download progress body")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	data := &extractors.Data{
		URL:   server.URL,
		Title: "clip",
		Type:  extractors.DataTypeVideo,
		Streams: map[string]*extractors.Stream{
			"default": {
				Parts: []*extractors.Part{
					{URL: server.URL, Size: int64(len(body)), Ext: "bin"},
				},
			},
		},
	}
	data.FillUpStreamsData()

	var events []ProgressEvent
	downloader := New(Options{
		Silent:         true,
		OutputPath:     t.TempDir(),
		ThreadNumber:   1,
		RetryTimes:     1,
		FileNameLength: 255,
		OnProgress: func(event ProgressEvent) {
			events = append(events, event)
		},
	})

	if err := downloader.Download(data); err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if len(events) == 0 {
		t.Fatal("Download() emitted no progress events")
	}
	last := events[len(events)-1]
	if last.Phase != ProgressFinished {
		t.Fatalf("last progress phase = %q, want %q", last.Phase, ProgressFinished)
	}
	if last.Current != int64(len(body)) || last.Total != int64(len(body)) || last.Percent != 1 {
		t.Fatalf("last progress = current %d total %d percent %.2f", last.Current, last.Total, last.Percent)
	}
}
