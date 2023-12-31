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

// CnfCertificationSuiteRunSpec defines the desired state of CnfCertificationSuiteRun
type CnfCertificationSuiteRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LabelsFilter holds the labels filter/expression of the test cases we want to run.
	LabelsFilter string `json:"labelsFilter"`
	// LogLevel sets the CNF Certification Suite log level (TNF_LOG_LEVEL)
	LogLevel string `json:"logLevel"`

	// Total timeout for the CNF Cert Suite to run.
	TimeOut string `json:"timeout"`
	// ConfigMapName holds the cnf certification suite yaml config.
	ConfigMapName string `json:"configMapName"`
	// PreflightSecretName holds the secret name for preflight's dockerconfig.
	PreflightSecretName string `json:"preflightSecretName"`
}

// CnfCertificationSuiteRunStatus defines the observed state of CnfCertificationSuiteRun
type CnfCertificationSuiteRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase holds the current phase of the CNF Certification Suite run.
	Phase string `json:"phase"`
	// Report Name of the CnfCertificationSuiteReport that has been created.
	ReportName string `json:"reportName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="CnfCertificationSuiteRun current status"

// CnfCertificationSuiteRun is the Schema for the cnfcertificationsuiteruns API
type CnfCertificationSuiteRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CnfCertificationSuiteRunSpec   `json:"spec,omitempty"`
	Status CnfCertificationSuiteRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CnfCertificationSuiteRunList contains a list of CnfCertificationSuiteRun
type CnfCertificationSuiteRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CnfCertificationSuiteRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CnfCertificationSuiteRun{}, &CnfCertificationSuiteRunList{})
}
