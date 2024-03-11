package definitions

const (
	CnfCertificationSuiteRunStatusPhaseCertSuiteDeploying     = "CertSuiteDeploying"
	CnfCertificationSuiteRunStatusPhaseCertSuiteDeployFailure = "CertSuiteDeployFailure"
	CnfCertificationSuiteRunStatusCertSuiteRunning            = "CertSuiteRunning"
	CnfCertificationSuiteRunStatusPhaseJobFinished            = "CertSuiteFinished"
	CnfCertificationSuiteRunStatusPhaseJobError               = "CertSuiteError"
)

const (
	CnfCertPodNamePrefix             = "cnf-job-run"
	CnfCertSuiteSidecarContainerName = "cnf-certsuite-sidecar"
	CnfCertSuiteContainerName        = "cnf-certsuite"

	CnfCertSuiteBaseFolder      = "/cnf-certsuite"
	CnfCnfCertSuiteConfigFolder = CnfCertSuiteBaseFolder + "/config/suite"
	CnfPreflightConfigFolder    = CnfCertSuiteBaseFolder + "/config/preflight"
	CnfCertSuiteResultsFolder   = CnfCertSuiteBaseFolder + "/results"

	CnfCertSuiteConfigFilePath    = CnfCnfCertSuiteConfigFolder + "/tnf_config.yaml"
	PreflightDockerConfigFilePath = CnfPreflightConfigFolder + "/preflight_dockerconfig.json"

	SideCarResultsFolderEnvVar = "TNF_RESULTS_FOLDER"

	SideCarImageEnvVar = "SIDECAR_APP_IMG"

	SideCarShowAllResultsLogsEnvVar = "SHOW_ALL_RESULTS_LOGS"
)
