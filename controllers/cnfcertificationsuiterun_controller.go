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

var certificationRuns map[certificationRun]bool
var cnfRunJobId int

//+kubebuilder:rbac:groups=cnf-certifications.redhat.com,resources=cnfcertificationsuiteruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cnf-certifications.redhat.com,resources=cnfcertificationsuiteruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cnf-certifications.redhat.com,resources=cnfcertificationsuiteruns/finalizers,verbs=update

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
		logrus.Errorf("Failed to fetch CnfCertificationSuiteRun CR %v", req.NamespacedName)

		logrus.Infof("Checking current running certification job for CnfCertificationSuiteRun %v", req.Namespace)
		if certificationRuns[reqCertificationRun] {
			logrus.Infof("There is a running Certification Suite job: cancelling it.")

			delete(certificationRuns, reqCertificationRun)
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if certificationRuns[reqCertificationRun] {
		logrus.Infof("There's a certification job running already. Ignoring changes in CnfCertificationSuiteRun %v", reqCertificationRun)
		return ctrl.Result{}, nil
	}

	logrus.Infof("New CNF Certification Job run requested: %v", reqCertificationRun)
	certificationRuns[reqCertificationRun] = true

	// Launch the tnf pod
	cnfRunJobId++
	cnfCertJobPod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("cnf-job-run-%d", cnfRunJobId),
			Namespace: "cnf-certification-operator"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "cnf-cert-suite",
					Image:   "quay.io/greyerof/tests:cnfsuiteopv2",
					Command: []string{"./run-cnf-suites.sh"},
					Args:    []string{"-l", "observability"},
					Env: []corev1.EnvVar{
						{
							Name:  "TNF_LOG_LEVEL",
							Value: "trace"},
						{
							Name:  "PFLT_DOCKERCONFIG",
							Value: "/usr/tnf/preflight.dummy"},
					},
					ImagePullPolicy: "Always",
				},
			},
		},
		Status: corev1.PodStatus{},
	}

	err := r.Create(ctx, &cnfCertJobPod)
	if err != nil {
		log.Log.Error(err, "Failed to create CNF Cert job")
		return ctrl.Result{}, nil
	}

	// var friendFound bool
	// if err := r.List(ctx, &podList); err != nil {
	// 	log.Error(err, "unable to list pods")
	// } else {
	// 	for _, item := range podList.Items {
	// 		if item.GetName() == foo.Spec.Name {
	// 			log.Info("pod linked to a foo custom resource found", "name", item.GetName())
	// 			friendFound = true
	// 		}
	// 	}
	// }

	// // Update Foo' happy status
	// foo.Status.Happy = friendFound
	// if err := r.Status().Update(ctx, &foo); err != nil {
	// 	log.Error(err, "unable to update foo's happy status", "status", friendFound)
	// 	return ctrl.Result{}, err
	// }
	// log.Info("foo's happy status updated", "status", friendFound)

	// log.Info("foo custom resource reconciled")

	// return ctrl.Result{}, nil

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CnfCertificationSuiteRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logrus.Infof("Setting up CnfCertificationSuiteRunReconciler's manager.")
	certificationRuns = map[certificationRun]bool{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cnfcertificationsv1alpha1.CnfCertificationSuiteRun{}).
		Complete(r)
}
