package definitions

const (
	CnfCertPodNamePrefix = "cnf-job-run"

	CnfCertSuiteBaseFolder      = "/cnf-certsuite"
	CnfCnfCertSuiteConfigFolder = CnfCertSuiteBaseFolder + "/config/suite"
	CnfPreflightConfigFolder    = CnfCertSuiteBaseFolder + "/config/preflight"
	CnfCertSuiteResultsFolder   = CnfCertSuiteBaseFolder + "/results"

	CnfCertSuiteConfigFilePath    = CnfCnfCertSuiteConfigFolder + "/tnf_config.yaml"
	PreflightDockerConfigFilePath = CnfPreflightConfigFolder + "/preflight_dockerconfig.json"

	SideCarResultsFolderEnvVar = "TNF_RESULTS_FOLDER"
)
