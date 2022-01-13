package metav1

import "errors"

func (r RunnerEngineList) isValid(e RunnerEngine) bool {
	validEngine := false
	for _, engine := range RunnerEngineListAll {
		if e == engine {
			validEngine = true
			break
		}
	}

	return validEngine
}

func (j *JobCreate) Validate() error {
	if j.Name == "" {
		return errors.New("name is required")
	}

	if j.RunnerEngine == "" {
		return errors.New("runner engine is required")
	}

	validEngine := RunnerEngineListAll.isValid(j.RunnerEngine)

	if !validEngine {
		return errors.New("invalid engine")
	}

	if j.JavaScriptBundleSrc != "" && j.RunnerEngine != RunnerEngineChromium {
		return errors.New("javascript bundle source is available only for chromium engine")
	}

	return nil
}

func (j *JobPatch) Validate() error {
	if j.RunnerEngine != "" {
		validEngine := RunnerEngineListAll.isValid(j.RunnerEngine)

		if !validEngine {
			return errors.New("invalid engine")
		}
	}

	if j.JavaScriptBundleSrc != "" && j.RunnerEngine != RunnerEngineChromium {
		return errors.New("javascript bundle source is available only for chromium engine")
	}

	return nil
}

func (req *RequestQueueCreateAPI) Validate() error {
	if req.RunID == "" {
		return errors.New("runner id is required")
	}

	if req.URL == "" {
		return errors.New("request queue url is required")
	}

	if req.Depth == 0 {
		return errors.New("request queue depth is required")
	}

	return nil
}
