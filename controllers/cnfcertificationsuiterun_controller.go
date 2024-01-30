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

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	cnfcertjob "github.com/greyerof/cnf-certification-operator/controllers/cnf-cert-job"
	"github.com/greyerof/cnf-certification-operator/controllers/definitions"

	"github.com/sirupsen/logrus"
)

var sideCarImage string

// CnfCertificationSuiteRunReconciler reconciles a CnfCertificationSuiteRun object
type CnfCertificationSuiteRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type certificationRun struct {
	name      string
	namespace string
}

var (
	// certificationRuns maps a certificationRun to a pod name
	certificationRuns map[certificationRun]string
	// Holds an autoincremental CNF Cert Suite pod id
	cnfRunPodID int
)

const (
	checkInterval              = 5 * time.Second
	defaultCnfCertSuiteTimeout = time.Hour
)

// +kubebuilder:rbac:groups=cnf-certifications.redhat.com,namespace=cnf-certification-operator,resources=cnfcertificationsuiteruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cnf-certifications.redhat.com,namespace=cnf-certification-operator,resources=cnfcertificationsuiteruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cnf-certifications.redhat.com,namespace=cnf-certification-operator,resources=cnfcertificationsuiteruns/finalizers,verbs=update

// +kubebuilder:rbac:groups=cnf-certifications.redhat.com,namespace=cnf-certification-operator,resources=cnfcertificationsuitereports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cnf-certifications.redhat.com,namespace=cnf-certification-operator,resources=cnfcertificationsuitereports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cnf-certifications.redhat.com,namespace=cnf-certification-operator,resources=cnfcertificationsuitereports/finalizers,verbs=update

// +kubebuilder:rbac:groups="",namespace=cnf-certification-operator,resources=pods,verbs=get;list;watch;create;update;patch;delete

func ignoreUpdatePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR
			return false
		},
	}
}

// Updates CnfCertificationSuiteRun.Status.Phase corresponding to a given status
func (r *CnfCertificationSuiteRunReconciler) updateJobPhaseStatus(cnfrun *cnfcertificationsv1alpha1.CnfCertificationSuiteRun, status string) {
	cnfrun.Status.Phase = status
	err := r.Status().Update(context.Background(), cnfrun)
	if err != nil {
		logrus.Errorf("Error found while updating CnfCertificationSuiteRun's status: %s", err)
	}
}

func (r *CnfCertificationSuiteRunReconciler) getJobRunTimeThreshold(timeoutStr string) time.Duration {
	jobRunTimeThreshold, err := time.ParseDuration(timeoutStr)
	if err != nil {
		logrus.Info("Couldn't extarct job run timeout, setting default timeout.")
		return defaultCnfCertSuiteTimeout
	}
	return jobRunTimeThreshold
}

func (r *CnfCertificationSuiteRunReconciler) waitForCnfCertJobPodToComplete(ctx context.Context, namespace string, cnfCertJobPod *corev1.Pod, jobRunTimeThreshold time.Duration) {
	cnfCertJobNamespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      cnfCertJobPod.Name,
	}

	startTime := time.Now()
	for {
		if time.Since(startTime) > jobRunTimeThreshold {
			logrus.Error("Time threshold reached, job did not complete")
			break
		}
		switch cnfCertJobPod.Status.Phase {
		case corev1.PodSucceeded:
			logrus.Info("Cnf job pod has completed successfully.")
			return
		case corev1.PodFailed:
			logrus.Info("Cnf job pod has completed with failure.")
			return
		default:
			logrus.Info("Cnf job pod is running. Current status: ", cnfCertJobPod.Status.Phase)
			time.Sleep(checkInterval)
		}
		err := r.Get(ctx, cnfCertJobNamespacedName, cnfCertJobPod)
		if err != nil {
			logrus.Error("Error found while getting cnf cert job pod: ", err)
		}
	}
}

func (r *CnfCertificationSuiteRunReconciler) getCertSuiteContainerExitStatus(cnfCertJobPod *corev1.Pod) int32 {
	for i := range cnfCertJobPod.Status.ContainerStatuses {
		containerStatus := &cnfCertJobPod.Status.ContainerStatuses[i]
		if containerStatus.Name == definitions.CnfCertSuiteContainerName {
			return containerStatus.State.Terminated.ExitCode
		}
	}
	return -1
}

func (r *CnfCertificationSuiteRunReconciler) handleEndOfCnfCertSuiteRun(ctx context.Context, namespace string, cnfCertJobPod *corev1.Pod, cnfrun *cnfcertificationsv1alpha1.CnfCertificationSuiteRun) {
	jobRunTimeThreshold := r.getJobRunTimeThreshold(cnfrun.Spec.TimeOut)
	r.waitForCnfCertJobPodToComplete(ctx, namespace, cnfCertJobPod, jobRunTimeThreshold)

	// cnf-cert-job has terminated - checking exit status of cert suite
	certSuiteExitStatus := r.getCertSuiteContainerExitStatus(cnfCertJobPod)
	if certSuiteExitStatus == 0 {
		r.updateJobPhaseStatus(cnfrun, "CertSuiteFinished")
		logrus.Info("CNF Cert job has finished running.")
	} else {
		r.updateJobPhaseStatus(cnfrun, "CertSuiteError")
		logrus.Info("CNF Cert job encountered an error. Exit status: ", certSuiteExitStatus)
	}

	r.updateRunCrStatusReportName(ctx, namespace, fmt.Sprintf("%s-report", cnfCertJobPod.Name), cnfrun)
}

func (r *CnfCertificationSuiteRunReconciler) waitForReportToBeCreated(ctx context.Context, namespace, reportName string) {
	reportNamespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      reportName,
	}
	startTime := time.Now()
	var cnfreport cnfcertificationsv1alpha1.CnfCertificationSuiteReport
	for err := r.Get(ctx, reportNamespacedName, &cnfreport); err != nil; {
		if time.Since(startTime) > defaultCnfCertSuiteTimeout {
			logrus.Error("Time threshold reached, report is not found")
			break
		}
		logrus.Infof("Waiting for %s to be created...", reportNamespacedName.Name)
		time.Sleep(checkInterval)
		err = r.Get(ctx, reportNamespacedName, &cnfreport)
	}
	logrus.Infof("%s has been created", reportNamespacedName.Name)
}

func (r *CnfCertificationSuiteRunReconciler) updateRunCrStatusReportName(ctx context.Context, namespace, reportName string, cnfrun *cnfcertificationsv1alpha1.CnfCertificationSuiteRun) {
	r.waitForReportToBeCreated(ctx, namespace, reportName)
	cnfrun.Status.ReportName = reportName
	err := r.Status().Update(context.Background(), cnfrun)
	if err != nil {
		logrus.Errorf("Error found while updating CnfCertificationSuiteRun's status: %s", err)
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CnfCertificationSuiteRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CnfCertificationSuiteRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	logrus.Infof("Reconciling CnfCertificationSuiteRun CRD.")

	reqCertificationRun := certificationRun{name: req.Name, namespace: req.Namespace}
	var cnfrun cnfcertificationsv1alpha1.CnfCertificationSuiteRun
	if getErr := r.Get(ctx, req.NamespacedName, &cnfrun); getErr != nil {
		logrus.Infof("CnfCertificationSuiteRun CR %s (ns %s) not found.", req.Name, req.NamespacedName)
		if podName, exist := certificationRuns[reqCertificationRun]; exist {
			logrus.Infof("CnfCertificationSuiteRun has been deleted. Removing the associated CNF Cert job pod %v", podName)
			deleteErr := r.Delete(context.TODO(), &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: req.Namespace}})
			if deleteErr != nil {
				logrus.Errorf("Failed to remove CNF Cert Job pod %s in namespace %s: %v", req.Name, req.Namespace, deleteErr)
			}
			delete(certificationRuns, reqCertificationRun)
		}
		return ctrl.Result{}, client.IgnoreNotFound(getErr)
	}

	if podName, exist := certificationRuns[reqCertificationRun]; exist {
		logrus.Infof("There's a certification job pod=%v running already. Ignoring changes in CnfCertificationSuiteRun %v", podName, reqCertificationRun)
		return ctrl.Result{}, nil
	}

	logrus.Infof("New CNF Certification Job run requested: %v", reqCertificationRun)

	cnfRunPodID++
	podName := fmt.Sprintf("%s-%d", definitions.CnfCertPodNamePrefix, cnfRunPodID)

	// Store the new run & associated CNF Cert pod name
	certificationRuns[reqCertificationRun] = podName

	logrus.Infof("Running CNF Certification Suite container (job id=%d) with labels %q, log level %q and timeout: %q",
		cnfRunPodID, cnfrun.Spec.LabelsFilter, cnfrun.Spec.LogLevel, cnfrun.Spec.TimeOut)

	// Launch the pod with the CNF Cert Suite container plus the sidecar container to fetch the results.
	r.updateJobPhaseStatus(&cnfrun, "CreatingCertSuiteJob")
	logrus.Info("Creating CNF Cert job pod")
	config := cnfcertjob.NewConfig(
		podName,
		req.Namespace,
		cnfrun.Name,
		cnfrun.Spec.LabelsFilter,
		cnfrun.Spec.LogLevel,
		cnfrun.Spec.TimeOut,
		cnfrun.Spec.ConfigMapName,
		cnfrun.Spec.PreflightSecretName,
		sideCarImage)
	cnfCertJobPod := cnfcertjob.New(config)

	err := r.Create(ctx, cnfCertJobPod)
	if err != nil {
		log.Log.Error(err, "Failed to create CNF Cert job pod")
		r.updateJobPhaseStatus(&cnfrun, "FailedToDeployCertSuitePod")
		return ctrl.Result{}, nil
	}
	r.updateJobPhaseStatus(&cnfrun, "RunningCertSuite")
	logrus.Info("Running CNF Cert job")

	go r.handleEndOfCnfCertSuiteRun(ctx, req.NamespacedName.Namespace, cnfCertJobPod, &cnfrun)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CnfCertificationSuiteRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logrus.Infof("Setting up CnfCertificationSuiteRunReconciler's manager.")
	certificationRuns = map[certificationRun]string{}

	var found bool
	sideCarImage, found = os.LookupEnv(definitions.SideCarImageEnvVar)
	if !found {
		return fmt.Errorf("sidecar app img env var %q not found", definitions.SideCarImageEnvVar)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cnfcertificationsv1alpha1.CnfCertificationSuiteRun{}).
		WithEventFilter(ignoreUpdatePredicate()).
		Complete(r)
}
