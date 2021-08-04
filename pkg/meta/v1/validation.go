package v1

import "errors"

func (j *JobCreate) Validate() error {
	if j.Name == "" {
		return errors.New("name is required")
	}

	if j.RunnerEngine == "" {
		return errors.New("runner engine is required")
	}

	validEngine := RunnerEngineAll.isValid(j.RunnerEngine)

	if !validEngine {
		return errors.New("invalid engine")
	}

	if j.JavaScriptBundleSrc != "" && j.RunnerEngine != RunnerEngineChromium {
		return errors.New("(JobCreate): javascript bundle source is available only for chromium engine")
	}

	return nil
}

func (j *JobPatch) Validate() error {
	if j.RunnerEngine != "" {
		validEngine := RunnerEngineAll.isValid(j.RunnerEngine)

		if !validEngine {
			return errors.New("invalid engine")
		}
	}

	if j.JavaScriptBundleSrc != "" && j.RunnerEngine != RunnerEngineChromium {
		return errors.New("(JobCreate): javascript bundle source is available only for chromium engine")
	}

	return nil
}
