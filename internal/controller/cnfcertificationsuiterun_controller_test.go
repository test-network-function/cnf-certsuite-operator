package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	cnfcertificationsv1alpha1 "github.com/test-network-function/cnf-certsuite-operator/api/v1alpha1"
	"github.com/test-network-function/cnf-certsuite-operator/internal/controller/definitions"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Creates a reconciler with a fake client to mock API calls.
func mockReconciler(objs []runtime.Object) *CnfCertificationSuiteRunReconciler {
	s := scheme.Scheme
	s.AddKnownTypes(cnfcertificationsv1alpha1.GroupVersion, &cnfcertificationsv1alpha1.CnfCertificationSuiteRun{})

	runCR := &cnfcertificationsv1alpha1.CnfCertificationSuiteRun{}
	cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).WithStatusSubresource(runCR).Build()

	return &CnfCertificationSuiteRunReconciler{Client: cl, Scheme: s}
}

func Test_getJobRunTimeThreshold(t *testing.T) {
	tests := []struct {
		name       string
		timeoutStr string
		want       time.Duration
	}{
		{ // Test case #1 - Pass with given timeout
			name:       "Set timeout",
			timeoutStr: "2h",
			want:       2 * time.Hour,
		},
		{ // Test case #2 - Pass with default timeout as timeoutStr is an empty string
			name:       "Empty timeout",
			timeoutStr: "",
			want:       time.Hour,
		},
	}

	for _, tc := range tests {
		assert.Equal(t, getJobRunTimeThreshold(tc.timeoutStr), tc.want)
	}
}

func Test_getCertSuiteContainerExitStatus(t *testing.T) {
	tests := []struct {
		name           string
		certSuitePod   *corev1.Pod
		wantExitStatus int32
		wantError      error
	}{
		{ // Test case #1 - Pass with exit status 0
			name: "Container cnf-certsuite has 0 exit code",
			certSuitePod: &corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "cnf-certsuite",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 0,
								},
							},
						},
					},
				},
			},
			wantExitStatus: 0,
			wantError:      nil,
		},
		{ // Test case #2 - Pass with exit status -1
			name: "Container cnf-certsuite wasn't find in pod",
			certSuitePod: &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "cnf-job-sample",
					Namespace: "cnf-certsuite-operator",
				},
				Status: corev1.PodStatus{},
			},
			wantExitStatus: 0,
			wantError:      fmt.Errorf("failed to get cert suite exit status: container not found in pod cnf-job-sample (ns cnf-certsuite-operator)"),
		},
	}
	for _, tc := range tests {
		gotExitStatus, gotErr := getCertSuiteContainerExitStatus(tc.certSuitePod)
		assert.Equal(t, gotErr, tc.wantError)
		if gotErr == nil { // no need to check exit status if an error has occurred
			assert.Equal(t, gotExitStatus, tc.wantExitStatus)
		}
	}
}

func TestCnfCertificationSuiteRunReconciler_waitForCertSuitePodToComplete(t *testing.T) {
	tests := []struct {
		name               string
		timeOut            time.Duration
		phase              corev1.PodPhase
		wantExitStatusCode int32
		wantError          error
	}{
		{ // Test case #1 - Pass with exit status 0
			name:               "Pass with Pod succeed phase",
			timeOut:            time.Hour,
			phase:              corev1.PodSucceeded,
			wantExitStatusCode: 0,
			wantError:          nil,
		},
		{ // Test case #2 - Fail, Pod stuck in running phase
			name:               "Failed with Pod running phase",
			timeOut:            10 * time.Second,
			phase:              corev1.PodRunning,
			wantExitStatusCode: 0,
			wantError:          fmt.Errorf("timeout (10s) reached while waiting for cert suite pod cnf-certsuite-operator/cnf-job-sample to finish"),
		},
	}

	certSuitePodNamespacedName := types.NamespacedName{
		Name:      "cnf-job-sample",
		Namespace: "cnf-certsuite-operator",
	}
	for _, tc := range tests {
		certSuitePod := &corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name:      certSuitePodNamespacedName.Name,
				Namespace: certSuitePodNamespacedName.Namespace,
			},
			Status: corev1.PodStatus{
				Phase: tc.phase,
			},
		}

		r := mockReconciler([]runtime.Object{certSuitePod})

		gotExitStatusCode, gotError := r.waitForCertSuitePodToComplete(certSuitePodNamespacedName, tc.timeOut)
		assert.Equal(t, gotError, tc.wantError)
		if gotError == nil { // no need to check exit status if an error has occurred
			assert.Equal(t, gotExitStatusCode, tc.wantExitStatusCode)
		}
	}
}

func TestCnfCertificationSuiteRunReconciler_updateStatus(t *testing.T) {
	tests := []struct {
		name                string
		statusSetterFn      func(currStatus *cnfcertificationsv1alpha1.CnfCertificationSuiteRunStatus)
		runCRNamespacedName types.NamespacedName
		wantErr             bool
	}{
		{ // Test case #1 - Pass with exit status 0
			name: "Pass when updating phase",
			statusSetterFn: func(currStatus *cnfcertificationsv1alpha1.CnfCertificationSuiteRunStatus) {
				currStatus.Phase = definitions.CnfCertificationSuiteRunStatusPhaseJobFinished
			},
			runCRNamespacedName: types.NamespacedName{
				Name:      "cnf-run-sample",
				Namespace: "cnf-certsuite-operator",
			},
			wantErr: false,
		},
		{ // Test case #1 - Fail, error = cnfcertificationsuiteruns.cnf-certifications.redhat.com "" not found
			name: "Fail updating phase",
			statusSetterFn: func(currStatus *cnfcertificationsv1alpha1.CnfCertificationSuiteRunStatus) {
				currStatus.Phase = definitions.CnfCertificationSuiteRunStatusPhaseJobFinished
			},
			runCRNamespacedName: types.NamespacedName{},
			wantErr:             true,
		},
	}

	runCR := &cnfcertificationsv1alpha1.CnfCertificationSuiteRun{
		ObjectMeta: v1.ObjectMeta{
			Name:      "cnf-run-sample",
			Namespace: "cnf-certsuite-operator",
		},
		Status: cnfcertificationsv1alpha1.CnfCertificationSuiteRunStatus{
			Phase: definitions.CnfCertificationSuiteRunStatusCertSuiteRunning,
		},
	}
	for _, tc := range tests {
		r := mockReconciler([]runtime.Object{runCR})

		if err := r.updateStatus(tc.runCRNamespacedName, tc.statusSetterFn); (err != nil) != tc.wantErr {
			t.Errorf("CnfCertificationSuiteRunReconciler.updateStatus() error = %v, wantErr %v", err, tc.wantErr)
		}
	}
}
