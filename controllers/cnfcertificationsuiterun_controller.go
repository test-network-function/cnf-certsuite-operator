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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	"github.com/greyerof/cnf-certification-operator/controllers/cnf-cert-job"
	"github.com/greyerof/cnf-certification-operator/controllers/definitions"

	"github.com/sirupsen/logrus"
)

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
	cnfRunPodId int
)

// +kubebuilder:rbac:groups="*",resources="*",verbs="*"
// +kubebuilder:rbac:urls="*",verbs="*"

// Updates CnfCertificationSuiteRun.Status.Phase correspomding to a given status
func (r *CnfCertificationSuiteRunReconciler) updateJobStatus(cnfrun *cnfcertificationsv1alpha1.CnfCertificationSuiteRun, status string) {
	cnfrun.Status.Phase = status
	r.Status().Update(context.Background(), cnfrun)
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

	reqCertificationRun := certificationRun{
		name:      req.Name,
		namespace: req.Namespace,
	}

	var cnfrun cnfcertificationsv1alpha1.CnfCertificationSuiteRun
	if err := r.Get(ctx, req.NamespacedName, &cnfrun); err != nil {
		logrus.Infof("CnfCertificationSuiteRun CR %s (ns %s) not found.", req.Name, req.NamespacedName)

		if podName, exist := certificationRuns[reqCertificationRun]; exist {
			logrus.Infof("CnfCertificationSuiteRun has been deleted. Removing the associated CNF Cert job pod %v", podName)

			err := r.Delete(context.TODO(), &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: req.Namespace}})
			if err != nil {
				logrus.Errorf("Failed to remove CNF Cert Job pod %s in namespace %s: %v", req.Name, req.Namespace, err)
			}

			delete(certificationRuns, reqCertificationRun)
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if podName, exist := certificationRuns[reqCertificationRun]; exist {
		logrus.Infof("There's a certification job pod=%v running already. Ignoring changes in CnfCertificationSuiteRun %v", podName, reqCertificationRun)
		return ctrl.Result{}, nil
	}

	logrus.Infof("New CNF Certification Job run requested: %v", reqCertificationRun)

	cnfRunPodId++
	podName := fmt.Sprintf("%s-%d", definitions.CnfCertPodNamePrefix, cnfRunPodId)

	// Store the new run & associated CNF Cert pod name
	certificationRuns[reqCertificationRun] = podName

	logrus.Infof("Running CNF Certification Suite container (job id=%d) with labels %q, log level %q and timeout: %q",
		cnfRunPodId, cnfrun.Spec.LabelsFilter, cnfrun.Spec.LogLevel, cnfrun.Spec.TimeOut)

	// Launch the pod with the CNF Cert Suite container plus the sidecar container to fetch the results.
	r.updateJobStatus(&cnfrun, "CreatingCertSuiteJob")
	config := cnfcertjob.NewConfig(cnfrun.Spec.LabelsFilter, cnfrun.Spec.LogLevel, cnfrun.Spec.ConfigMapName, cnfrun.Spec.PreflightSecretName)
	cnfCertJobPod := cnfcertjob.New(config, podName)

	r.updateJobStatus(&cnfrun, "RunningCertSuite")
	err := r.Create(ctx, cnfCertJobPod)
	if err != nil {
		log.Log.Error(err, "Failed to create CNF Cert job")
		r.updateJobStatus(&cnfrun, "CertSuiteError")
		return ctrl.Result{}, nil
	}

	r.updateJobStatus(&cnfrun, "CertSuiteFinished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CnfCertificationSuiteRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logrus.Infof("Setting up CnfCertificationSuiteRunReconciler's manager.")
	certificationRuns = map[certificationRun]string{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cnfcertificationsv1alpha1.CnfCertificationSuiteRun{}).
		Complete(r)
}
