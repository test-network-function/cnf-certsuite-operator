package cnfcertsuitereport

import (
	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	"github.com/greyerof/cnf-certification-operator/cnf-cert-sidecar/app/claim"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	ReportCrName           string
	Namespace              string
	CertSuiteConfigRunName string
	OcpVersion             string
	CnfCertSuiteVersion    string
	Cnf                    cnfcertificationsv1alpha1.CnfTargets
}

func addNamespacesToCnfSpecField(cnf *cnfcertificationsv1alpha1.CnfTargets, namespaces []string) {
	cnf.Namespaces = append(cnf.Namespaces, namespaces...)
}

func addPodsToCnfSpecField(cnf *cnfcertificationsv1alpha1.CnfTargets, pods []corev1.Pod) {
	for podIdx := 0; podIdx < len(pods); podIdx++ {
		pod := &pods[podIdx]
		podsContainers := &pod.Spec.Containers
		var containers []string
		for containerIdx := 0; containerIdx < len(*podsContainers); containerIdx++ {
			containers = append(containers, (*podsContainers)[containerIdx].Name)
		}
		cnf.Pods = append(cnf.Pods, cnfcertificationsv1alpha1.CnfPod{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Containers: containers,
		})
	}
}

func addOperatorsToCnfSpecField(cnf *cnfcertificationsv1alpha1.CnfTargets, csvs []claim.Metadata) {
	for _, csv := range csvs {
		cnf.Csvs = append(cnf.Csvs, cnfcertificationsv1alpha1.CnfResource{
			Name:      csv.Name,
			Namespace: csv.Namespace,
		})
	}
}

func addCrdsToCnfSpecField(cnf *cnfcertificationsv1alpha1.CnfTargets, crds []claim.Resource) {
	for _, crd := range crds {
		cnf.Crds = append(cnf.Crds, crd.Metadata.Name)
	}
}

func addNodesToCnfSpecField(cnf *cnfcertificationsv1alpha1.CnfTargets, nodes map[string]interface{}) {
	for nodeName := range nodes {
		cnf.Nodes = append(cnf.Nodes, nodeName)
	}
}

func addResourcesToCnfSpecField(cnfResourceField *[]cnfcertificationsv1alpha1.CnfResource, resources []claim.Resource) {
	for _, resource := range resources {
		*cnfResourceField = append(*cnfResourceField, cnfcertificationsv1alpha1.CnfResource{
			Name:      resource.Metadata.Name,
			Namespace: resource.Metadata.Namespace,
		})
	}
}

func getCnfTargetsFromClaim(claimcontent *claim.Schema) *cnfcertificationsv1alpha1.CnfTargets {
	var cnf cnfcertificationsv1alpha1.CnfTargets
	addNamespacesToCnfSpecField(&cnf, claimcontent.Claim.Configurations.NameSpaces)
	addNodesToCnfSpecField(&cnf, claimcontent.Claim.Nodes.NodeSummary)
	addPodsToCnfSpecField(&cnf, claimcontent.Claim.Configurations.Pods)
	addResourcesToCnfSpecField(&cnf.Deployments, claimcontent.Claim.Configurations.Deployments)
	addResourcesToCnfSpecField(&cnf.StatefulSets, claimcontent.Claim.Configurations.StatefulSets)
	addOperatorsToCnfSpecField(&cnf, claimcontent.Claim.Configurations.Csvs)
	addCrdsToCnfSpecField(&cnf, claimcontent.Claim.Configurations.Crds)
	addResourcesToCnfSpecField(&cnf.Services, claimcontent.Claim.Configurations.Services)
	addResourcesToCnfSpecField(&cnf.HelmChartReleases, claimcontent.Claim.Configurations.HelmChartReleases)
	return &cnf
}

func New(config *Config) *cnfcertificationsv1alpha1.CnfCertificationSuiteReport {
	return &cnfcertificationsv1alpha1.CnfCertificationSuiteReport{
		OcpVersion:          config.OcpVersion,
		CnfCertSuiteVersion: config.CnfCertSuiteVersion,
		CnfTargets:          config.Cnf,
	}
}

func SetRunCRStatus(runCR *cnfcertificationsv1alpha1.CnfCertificationSuiteRun, claimSchema *claim.Schema) {
	testSuiteResults := &claimSchema.Claim.Results
	results := []cnfcertificationsv1alpha1.TestCaseResult{}
	totalTests, passedTests, skippedTests, failedTests, erroredTests := 0, 0, 0, 0, 0
	for tcName := range *testSuiteResults {
		tcResult := (*testSuiteResults)[tcName]
		testCaseResult := cnfcertificationsv1alpha1.TestCaseResult{
			TestCaseName: tcName,
			Result:       tcResult.State,
		}

		switch tcResult.State {
		case "passed":
			passedTests++
		case "skipped":
			skippedTests++
			testCaseResult.Reason = tcResult.SkipReason
		case "failed":
			failedTests++
			testCaseResult.Reason = tcResult.FailureReason
			testCaseResult.Logs = tcResult.CapturedTestOutput
		case "error": //nolint:goconst
			erroredTests++
		}

		if runCR.Spec.ShowAllResultsLogs {
			testCaseResult.Logs = tcResult.CapturedTestOutput
		}
		totalTests++
		results = append(results, testCaseResult)
	}

	runCR.Status.Report = &cnfcertificationsv1alpha1.CnfCertificationSuiteReport{
		OcpVersion:          claimSchema.Claim.Versions.Ocp,
		CnfCertSuiteVersion: claimSchema.Claim.Versions.Tnf,
		CnfTargets:          *getCnfTargetsFromClaim(claimSchema),
		Results:             results,
		Summary: cnfcertificationsv1alpha1.CnfCertificationSuiteReportStatusSummary{
			Total:   totalTests,
			Passed:  passedTests,
			Skipped: skippedTests,
			Failed:  failedTests,
			Errored: erroredTests,
		},
	}

	switch {
	case erroredTests >= 1: // at least one test encountered an error
		runCR.Status.Report.Verdict = "error"
	case failedTests >= 1: // at least one failed test
		runCR.Status.Report.Verdict = "fail"
	case skippedTests == totalTests: // all tests were skipped
		runCR.Status.Report.Verdict = "skip"
	default: // all tests who ran have passed
		runCR.Status.Report.Verdict = "pass"
	}
}
