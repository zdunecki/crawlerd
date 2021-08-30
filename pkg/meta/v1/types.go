package v1

import (
	"fmt"
	"strings"
)

// TODO: add protobuf for structs

type RunnerStatus string

const (
	RunnerStatusQueued RunnerStatus = "Queued"

	RunnerStatusSuccessed RunnerStatus = "Successed"

	RunnerStatusFailed RunnerStatus = "Failed"

	RunnerStatusTimedOut RunnerStatus = "TimedOut"
)

type RunnerEngine string

const (
	RunnerEngineJavaScript RunnerEngine = "JavaScript"

	RunnerEngineChromium RunnerEngine = "Chromium"

	RunnerEngineCrawlBot RunnerEngine = "CrawlBot"
)

type RunnerEngineList []RunnerEngine

var RunnerEngineListAll RunnerEngineList = []RunnerEngine{RunnerEngineJavaScript, RunnerEngineChromium, RunnerEngineCrawlBot}

// RunnerInitialDepth is start depth value, should be counter from 1.
const RunnerInitialDepth = 1

type Runner struct {
	ID string `json:"id" bson:"_id"`

	RunAt int `json:"run_at" bson:"run_at"`

	EndAt int `json:"end_at" bson:"end_at"`

	Status RunnerStatus `json:"status" bson:"status"`

	// Depth current depth, not greater than RunnerUpCreate.MaxDepth.
	Depth uint `json:"depth" bson:"depth"`
}

type RunnerCreate struct {
	RunAt int `json:"run_at" bson:"run_at"`

	EndAt int `json:"end_at" bson:"end_at,omitempty"`

	Status RunnerStatus `json:"status" bson:"status"`

	Depth uint `json:"depth" bson:"depth"`
}

type RunnerPatch struct {
	EndAt int `json:"end_at" bson:"end_at,omitempty"`

	Status RunnerStatus `json:"status" bson:"status,omitempty"`

	Depth uint `json:"depth" bson:"depth"`
}

type RunnerUpCreate struct {
	// ID is an identifier helpful to find runner config/functions etc.
	ID string `json:"id"`

	// URL is a page url where runner should crawl.
	URL string `json:"url"`

	// ScrapeLinksPattern is a regexp pattern in which links should be scraped, useful for narrow down results.
	// If not defined, then match all.
	ScrapeLinksPattern string `json:"scrape_links_pattern"`

	// FollowLinks which links should be crawled deeper.
	// If not defined, then follow all links.
	// Applied only if MaxDepth is
	FollowLinks []*StringFilter `json:"follow_links"`

	// MaxDepth how deep FollowLinks should be, counted from first run.
	// Should be greater than or equal RunnerInitialDepth.
	// By default, if not specified then runner crawl just once, even if FollowLinks is defined.
	MaxDepth uint `json:"max_depth"`
}

type Job struct {
	ID string `json:"id" bson:"_id"`

	// URL is an url to crawled page
	URL string `json:"url" bson:"url"`

	// Name user-friendly name for a job.
	Name string `json:"name" bson:"name"`

	// RunnerEngine is a preconfigured environment in where crawler is running.
	RunnerEngine RunnerEngine `json:"runner_engine" bson:"runner_engine"`

	// JavaScriptBundleSrc is a path to bundled file in cloud-native storage like S3,GCS.
	// Currently, supported only by RunnerEngineChromium.
	JavaScriptBundleSrc string `json:"js_bundle_src" bson:"js_bundle_src"`

	// Runs is a value how many times Job was run
	Runs int `json:"runs" bson:"runs"`

	LastRunAt int `json:"last_run_at" bson:"last_run_at"`

	LastRunEndAt int `json:"last_run_end_at" bson:"last_run_end_at"`

	LastRunStatus RunnerStatus `json:"last_run_status" bson:"last_run_status"`
}

type JobCreate struct {
	Name string `json:"name" bson:"name"`

	RunnerEngine RunnerEngine `json:"runner_engine" bson:"runner_engine"`

	JavaScriptBundleSrc string `json:"js_bundle_src" bson:"js_bundle_src"`
}

type JobPatch struct {
	Name string `json:"name" bson:"name,omitempty"`

	URL string `json:"url" bson:"url"`

	RunnerEngine RunnerEngine `json:"runner_engine" bson:"runner_engine,omitempty"`

	JavaScriptBundleSrc string `json:"js_bundle_src" bson:"js_bundle_src,omitempty"`
}

func (j *JobPatch) ApplyJob(job *Job) {
	if j.Name == "" {
		j.Name = job.Name
	}

	if j.RunnerEngine == "" {
		j.RunnerEngine = job.RunnerEngine
	}

	if j.JavaScriptBundleSrc == "" {
		j.JavaScriptBundleSrc = job.JavaScriptBundleSrc
	}
}

// TODO: add status to avoid infinite
type RequestQueue struct {
	RunID string `json:"run_id"`

	URL string `json:"url" bson:"url"`
}

type RequestQueueCreate struct {
	RunID string `json:"run_id"`

	URL string `json:"url" bson:"url"`
}

type RequestQueueListFilter struct {
	RunID *StringFilter `json:"run_id" bson:"run_id"`

	URL *StringFilter `json:"url" bson:"url"`
}

type LinkURL string

func NewLinkURL(s string) LinkURL {
	return LinkURL(LinkURL(s).trim())
}

func (u LinkURL) trim() string {
	return strings.TrimSuffix(fmt.Sprintf("%s", u), "/")
}

type LinkNode struct {
	URL LinkURL `json:"url" bson:"url"`
}

type LinkNodeCreate struct {
	URL LinkURL `json:"url" bson:"url"`
}

// Deprecated: URL is not RequestQueueCreate
type URL struct {
	ID  int    `json:"id" bson:"id"`
	URL string `json:"url" bson:"url"`
	// Deprecated: Interval is not important now
	Interval int `json:"interval" bson:"interval"`
}

// Deprecated:
type History struct {
	Response        string  `json:"response" bson:"response"`
	CreatedAt       float64 `json:"created_at" bson:"created_at"`
	DurationSeconds float64 `json:"duration" bson:"duration"`
}

// Deprecated:
type CrawlURL struct {
	Id       int64  `json:"id"`
	Url      string `json:"url"`
	Interval int64  `json:"interval"`
	// Deprecated: Interval is not important now
	WorkerID string `json:"worker_id"`
}
