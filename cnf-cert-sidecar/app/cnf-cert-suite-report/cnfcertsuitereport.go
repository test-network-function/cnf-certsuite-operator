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
)

type Config struct {
	ReportCrName           string
	Namespace              string
	CertSuiteConfigRunName string
	OcpVersion             string
	CnfCertSuiteVersion    string
	Cnf                    cnfcertificationsv1alpha1.Cnf
}

func addNamespacesToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, namespaces *[]string) {
	cnf.Namespaces = append(cnf.Namespaces, *namespaces...)
}

func addPodsToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, pods *[]corev1.Pod) {
	for _, pod := range *pods {
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

func addOperatorsToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, csvs *[]claim.Metadata) {
	for _, csv := range *csvs {
		cnf.Deployments = append(cnf.Deployments, cnfcertificationsv1alpha1.CnfResource{
			Name:      csv.Name,
			Namespace: csv.Namespace,
		})
	}
}

func addCrdsToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, crds *[]claim.Resource) {
	for _, crd := range *crds {
		cnf.Crds = append(cnf.Crds, crd.Metadata.Name)
	}
}

func addNodesToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, nodes *map[string]interface{}) {
	for nodeName := range *nodes {
		cnf.Nodes = append(cnf.Nodes, nodeName)
	}
}

func addResourcesToCnfSpecField(cnf *cnfcertificationsv1alpha1.Cnf, resources *[]claim.Resource) {
	for _, resource := range *resources {
		cnf.Deployments = append(cnf.Deployments, cnfcertificationsv1alpha1.CnfResource{
			Name:      resource.Metadata.Name,
			Namespace: resource.Metadata.Namespace,
		})
	}
}

func newCnfSpecField(claimcontent *claim.Schema) *cnfcertificationsv1alpha1.Cnf {
	var cnf cnfcertificationsv1alpha1.Cnf
	addNamespacesToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.NameSpaces)
	addNodesToCnfSpecField(&cnf, &claimcontent.Claim.Nodes.NodeSummary)
	addPodsToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.Pods)
	addResourcesToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.Deployments)
	addResourcesToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.StatefulSets)
	addOperatorsToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.Csvs)
	addCrdsToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.Crds)
	addResourcesToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.Services)
	addResourcesToCnfSpecField(&cnf, &claimcontent.Claim.Configurations.HelmChartReleases)
	return &cnf
}

func NewConfig(claimContent *claim.Schema) *Config {
	certSuiteConfigRunName := os.Getenv(podNameEnvVar)
	reportCrName := fmt.Sprintf("%s-report", certSuiteConfigRunName)
	cnf := newCnfSpecField(claimContent)

	return &Config{
		ReportCrName:           reportCrName,
		Namespace:              os.Getenv(podNamespaceEnvVar),
		CertSuiteConfigRunName: certSuiteConfigRunName,
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
		switch tcResult.State {
		case "passed":
			passedTests++
		case "skipped":
			skippedTests++
		case "failed":
			failedTests++
		case "error":
			erroredTests++
		}
		totalTests++
		results = append(results, cnfcertificationsv1alpha1.TestCaseResult{
			TestCaseName: tcName,
			Result:       tcResult.State,
			Reason:       tcResult.FailureReason,
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

	if erroredTests >= 1 { // at least one test encountered an error
		cnfCertSuiteReport.Status.Verdict = "error"
	} else if failedTests >= 1 { // at least one failed test
		cnfCertSuiteReport.Status.Verdict = "fail"
	} else if skippedTests == totalTests { // all tests were skipped
		cnfCertSuiteReport.Status.Verdict = "skip"
	} else { // all tests who ran have passed
		cnfCertSuiteReport.Status.Verdict = "pass"
	}
}
