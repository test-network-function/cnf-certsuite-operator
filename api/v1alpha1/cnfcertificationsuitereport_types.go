/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CnfPod struct {
	Name       string   `json:"name,omitempty"`
	Namespace  string   `json:"namespace,omitempty"`
	Containers []string `json:"containers,omitempty"`
}

type CnfResource struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type Cnf struct {
	Namespaces        []string      `json:"namespaces,omitempty"`
	Nodes             []string      `json:"nodes,omitempty"`
	Pods              []CnfPod      `json:"pods,omitempty"`
	Deployments       []CnfResource `json:"deployments,omitempty"`
	StatefulSets      []CnfResource `json:"statefulSets,omitempty"`
	Csvs              []CnfResource `json:"csvs,omitempty"`
	Crds              []string      `json:"crds,omitempty"`
	Services          []CnfResource `json:"services,omitempty"`
	HelmChartReleases []CnfResource `json:"helmChartReleases,omitempty"`
}

// CnfCertificationSuiteReportSpec defines the desired state of CnfCertificationSuiteReport
type CnfCertificationSuiteReportSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	CertSuiteConfigRunName string `json:"certSuiteConfigRunName"`
	OcpVersion             string `json:"ocpVersion"`
	CnfCertSuiteVersion    string `json:"cnfCertSuiteVersion"`
	Cnf                    Cnf    `json:"cnf,omitempty"`
}

// TestCaseResult holds a test case result
type TestCaseResult struct {
	TestCaseName string `json:"testCaseName"`
	Result       string `json:"result"`
	Reason       string `json:"reason,omitempty"`
	Logs         string `json:"logs,omitempty"`
}

type CnfCertificationSuiteReportStatusSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
	Errored int `json:"errored"`
}

// CnfCertificationSuiteReportStatus defines the observed state of CnfCertificationSuiteReport
type CnfCertificationSuiteReportStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Verdict string                                   `json:"verdict"`
	Summary CnfCertificationSuiteReportStatusSummary `json:"summary"`
	Results []TestCaseResult                         `json:"results"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CnfCertificationSuiteReport is the Schema for the cnfcertificationsuitereports API
type CnfCertificationSuiteReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CnfCertificationSuiteReportSpec   `json:"spec,omitempty"`
	Status CnfCertificationSuiteReportStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CnfCertificationSuiteReportList contains a list of CnfCertificationSuiteReport
type CnfCertificationSuiteReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CnfCertificationSuiteReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CnfCertificationSuiteReport{}, &CnfCertificationSuiteReportList{})
}
