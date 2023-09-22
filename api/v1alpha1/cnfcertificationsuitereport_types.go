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

// CnfCertificationSuiteReportSpec defines the desired state of CnfCertificationSuiteReport
type CnfCertificationSuiteReportSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Results holds the result of each test case
	Results []TestCaseResult `json:"results"`
}

// TestCaseResult holds a test case result
type TestCaseResult struct {
	TestCaseName string `json:"testCaseName"`
	Result       string `json:"result"`
}

// CnfCertificationSuiteReportStatus defines the observed state of CnfCertificationSuiteReport
type CnfCertificationSuiteReportStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
