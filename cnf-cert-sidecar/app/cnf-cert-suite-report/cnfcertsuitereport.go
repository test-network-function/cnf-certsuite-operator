package cnfcertsuitereport

import (
	"fmt"
	"os"

	"github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	"github.com/greyerof/cnf-certification-operator/cnf-cert-sidecar/app/claim"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podNameEnvVar      = "MY_POD_NAME"
	podNamespaceEnvVar = "MY_POD_NAMESPACE"
	cnfRunNameEnvVar   = "CNF_RUN_NAME"
)

type Config struct {
	ReportCrName           string
	Namespace              string
	CertSuiteConfigRunName string
	OcpVersion             string
	CnfCertSuiteVersion    string
	Cnf                    cnfcertificationsv1alpha1.Cnf
}

func addNamespacesToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, namespaces []string) {
	cnf.Namespaces = append(cnf.Namespaces, namespaces...)
}

func addPodsToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, pods []corev1.Pod) {
	for _, pod := range pods {
		var containers []string
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Name)
		}
		cnf.Pods = append(cnf.Pods, cnfcertificationsv1alpha1.CnfPod{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Containers: containers,
		})
	}
}

func addOperatorsToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, csvs []claim.Metadata) {
	for _, csv := range csvs {
		cnf.Csvs = append(cnf.Csvs, cnfcertificationsv1alpha1.CnfResource{
			Name:      csv.Name,
			Namespace: csv.Namespace,
		})
	}
}

func addCrdsToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, crds []claim.Resource) {
	for _, crd := range crds {
		cnf.Crds = append(cnf.Crds, crd.Metadata.Name)
	}
}

func addNodesToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, nodes map[string]interface{}) {
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

func newCnfSpecField(claimcontent *claim.Schema) *cnfcertificationsv1alpha1.Cnf {
	var cnf cnfcertificationsv1alpha1.Cnf
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

func NewConfig(claimContent *claim.Schema) *Config {
	reportCrName := fmt.Sprintf("%s-report", os.Getenv(podNameEnvVar))
	cnf := newCnfSpecField(claimContent)

	return &Config{
		ReportCrName:           reportCrName,
		Namespace:              os.Getenv(podNamespaceEnvVar),
		CertSuiteConfigRunName: os.Getenv(cnfRunNameEnvVar),
		OcpVersion:             claimContent.Claim.Versions.Ocp,
		CnfCertSuiteVersion:    claimContent.Claim.Versions.Tnf,
		Cnf:                    *cnf,
	}
}

func New(config *Config) *cnfcertificationsv1alpha1.CnfCertificationSuiteReport {
	return &cnfcertificationsv1alpha1.CnfCertificationSuiteReport{
		ObjectMeta: metav1.ObjectMeta{Name: config.ReportCrName, Namespace: config.Namespace},
		Spec: cnfcertificationsv1alpha1.CnfCertificationSuiteReportSpec{
			CertSuiteConfigRunName: config.CertSuiteConfigRunName,
			OcpVersion:             config.OcpVersion,
			CnfCertSuiteVersion:    config.CnfCertSuiteVersion,
			Cnf:                    config.Cnf,
		},
		Status: cnfcertificationsv1alpha1.CnfCertificationSuiteReportStatus{},
	}
}

func UpdateStatus(cnfCertSuiteReport *cnfcertificationsv1alpha1.CnfCertificationSuiteReport,
	testSuiteResults *claim.TestSuiteResults) {
	results := []cnfcertificationsv1alpha1.TestCaseResult{}
	totalTests, passedTests, skippedTests, failedTests, erroredTests := 0, 0, 0, 0, 0
	for tcName, tcResult := range *testSuiteResults {
		reason := ""
		switch tcResult.State {
		case "passed":
			passedTests++
		case "skipped":
			skippedTests++
			reason = tcResult.SkipReason
		case "failed":
			failedTests++
			reason = tcResult.FailureReason
		case "error":
			erroredTests++
		}
		totalTests++
		results = append(results, cnfcertificationsv1alpha1.TestCaseResult{
			TestCaseName: tcName,
			Result:       tcResult.State,
			Reason:       reason,
			Logs:         tcResult.CapturedTestOutput,
		})
	}
	cnfCertSuiteReport.Status.Results = results
	cnfCertSuiteReport.Status.Summary = v1alpha1.CnfCertificationSuiteReportStatusSummary{
		Total:   totalTests,
		Passed:  passedTests,
		Skipped: skippedTests,
		Failed:  failedTests,
		Errored: erroredTests,
	}

	switch {
	case erroredTests >= 1: // at least one test encountered an error
		cnfCertSuiteReport.Status.Verdict = "error"
	case failedTests >= 1: // at least one failed test
		cnfCertSuiteReport.Status.Verdict = "fail"
	case skippedTests == totalTests: // all tests were skipped
		cnfCertSuiteReport.Status.Verdict = "skip"
	default: // all tests who ran have passed
		cnfCertSuiteReport.Status.Verdict = "pass"
	}
}
