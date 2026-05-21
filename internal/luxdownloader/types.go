package luxdownloader

import "time"

// ProgressPhase is the current high-level state of a download.
type ProgressPhase string

const (
	ProgressExtracting  ProgressPhase = "extracting"
	ProgressDownloading ProgressPhase = "downloading"
	ProgressMerging     ProgressPhase = "merging"
	ProgressSkipped     ProgressPhase = "skipped"
	ProgressFinished    ProgressPhase = "finished"
)

// ProgressEvent is emitted when extraction or byte-level download progress changes.
type ProgressEvent struct {
	Phase    ProgressPhase
	URL      string
	Title    string
	StreamID string
	FileName string
	Message  string
	Current  int64
	Total    int64
	Percent  float64
}

// ProgressFunc receives progress events from the downloader.
type ProgressFunc func(ProgressEvent)

// Options defines options used in downloading.
type Options struct {
	InfoOnly       bool
	Silent         bool
	Stream         string
	AudioOnly      bool
	Refer          string
	OutputPath     string
	OutputName     string
	FileNameLength int
	Caption        bool

	MultiThread  bool
	ThreadNumber int
	RetryTimes   int
	ChunkSizeMB  int

	UseAria2RPC bool
	Aria2Token  string
	Aria2Method string
	Aria2Addr   string

	OnProgress       ProgressFunc
	ProgressThrottle time.Duration
}

// Aria2RPCData defines the data structure of JSON RPC 2.0 info for Aria2.
type Aria2RPCData struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      string         `json:"id"`
	Method  string         `json:"method"`
	Params  [3]interface{} `json:"params"`
}

// Aria2Input is options for aria2.addUri.
type Aria2Input struct {
	Out    string   `json:"out"`
	Header []string `json:"header"`
}

// FilePartMeta defines the data structure of file meta info.
type FilePartMeta struct {
	Index float32
	Start int64
	End   int64
	Cur   int64
}
