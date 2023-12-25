package claim

const (
	supportedClaimFormatVersion = "v0.1.0"
)

const (
	TestCaseResultPassed  = "passed"
	TestCaseResultSkipped = "skipped"
	TestCaseResultFailed  = "failed"
)

type TestCaseResult struct {
	CapturedTestOutput string `json:"capturedTestOutput"`
	CatalogInfo        struct {
		BestPracticeReference string `json:"bestPracticeReference"`
		Description           string `json:"description"`
		ExceptionProcess      string `json:"exceptionProcess"`
		Remediation           string `json:"remediation"`
	} `json:"catalogInfo"`
	CategoryClassification map[string]string `json:"categoryClassification"`
	Duration               int               `json:"duration"`
	EndTime                string            `json:"endTime"`
	FailureLineContent     string            `json:"failureLineContent"`
	FailureLocation        string            `json:"failureLocation"`
	FailureReason          string            `json:"failureReason"`
	StartTime              string            `json:"startTime"`
	State                  string            `json:"state"`
	TestID                 struct {
		ID    string `json:"id"`
		Suite string `json:"suite"`
		Tags  string `json:"tags"`
	} `json:"testID"`
}

// Maps a test suite name to a list of TestCaseResult
type TestSuiteResults map[string]TestCaseResult

type Schema struct {
	Claim struct {
		Results  TestSuiteResults `json:"results"`
		Versions struct {
			ClaimFormat string `json:"claimFormat"`
		} `json:"versions"`
	} `json:"claim"`
}
