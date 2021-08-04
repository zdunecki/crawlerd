package v1

type RunnerStatus string

const (
	RunnerStatusSuccessed RunnerStatus = "Successed"

	RunnerStatusFailed RunnerStatus = "Failed"

	RunnerStatusTimedOut RunnerStatus = "TimedOut"
)

type RunnerEngine string

const (
	RunnerEngineJavaScript RunnerEngine = "JavaScript"

	RunnerEngineChromium RunnerEngine = "Chromium"
)

type RunnerEngineList []RunnerEngine

func (r RunnerEngineList) isValid(e RunnerEngine) bool {
	validEngine := false
	for _, engine := range RunnerEngineAll {
		if e == engine {
			validEngine = true
			break
		}
	}

	return validEngine
}

var RunnerEngineAll RunnerEngineList = []RunnerEngine{RunnerEngineJavaScript, RunnerEngineChromium}

type Job struct {
	ID string `json:"id" bson:"_id"`

	// Name user-friendly name for a job.
	Name string `json:"name" bson:"name"`

	// RunnerEngine is a preconfigured environment where crawler is running.
	RunnerEngine RunnerEngine `json:"runner_engine" bson:"runner_engine"`

	// JavaScriptBundleSrc is a path to bundled file in cloud-native storage like S3,GCS.
	// Currently, supported only by RunnerEngineChromium
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

type (
	URL struct {
		ID       int    `json:"id" bson:"id"`
		URL      string `json:"url" bson:"url"`
		Interval int    `json:"interval" bson:"interval"`
	}

	History struct {
		Response        string  `json:"response" bson:"response"`
		CreatedAt       float64 `json:"created_at" bson:"created_at"`
		DurationSeconds float64 `json:"duration" bson:"duration"`
	}

	Sequence struct {
		ObjectID string `bson:"_id"`
		ID       int    `bson:"id"`
	}

	// TODO: protobuf
	CrawlURL struct {
		Id       int64  `json:"id"`
		Url      string `json:"url"`
		Interval int64  `json:"interval"`
		WorkerID string `json:"worker_id"`
	}
)
