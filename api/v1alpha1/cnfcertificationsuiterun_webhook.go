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
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var cnfcertificationsuiterunlog = logf.Log.WithName("cnfcertificationsuiterun-resource")

var c client.Client

func (r *CnfCertificationSuiteRun) SetupWebhookWithManager(mgr ctrl.Manager) error {
	err := r.createClient()
	if err != nil {
		return err
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (r *CnfCertificationSuiteRun) createClient() error {
	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("error getting OpenShift config: %v", err)
	}

	c, err = client.New(kubeconfig, client.Options{})
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	return nil
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll
//+kubebuilder:webhook:path=/validate-cnf-certifications-redhat-com-v1alpha1-cnfcertificationsuiterun,mutating=false,failurePolicy=fail,sideEffects=None,groups=cnf-certifications.redhat.com,resources=cnfcertificationsuiteruns,verbs=create;update,versions=v1alpha1,name=vcnfcertificationsuiterun.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CnfCertificationSuiteRun{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CnfCertificationSuiteRun) ValidateCreate() error {
	cnfcertificationsuiterunlog.Info("validate create", "name", r.Name)

	configMap := &v1.ConfigMap{}
	preflightSecret := &v1.Secret{}

	// Validate config map name field
	err := c.Get(context.TODO(), types.NamespacedName{Name: r.Spec.ConfigMapName, Namespace: r.Namespace}, configMap)
	if err != nil {
		cnfcertificationsuiterunlog.Error(err, "CnfCertificationSuiteRun's config map name field is invalid",
			"config map name", r.Spec.ConfigMapName)
		return err
	}
	cnfcertificationsuiterunlog.Info("CnfCertificationSuiteRun's config map name field is valid", "config map name", configMap.Name)

	// Validate preflight secret name field
	err = c.Get(context.TODO(), types.NamespacedName{Name: r.Spec.PreflightSecretName, Namespace: r.Namespace}, preflightSecret)
	if err != nil {
		cnfcertificationsuiterunlog.Error(err, "CnfCertificationSuiteRun's preflight secret name field is invalid",
			"preflight secret name", r.Spec.PreflightSecretName)
		return err
	}
	cnfcertificationsuiterunlog.Info("CnfCertificationSuiteRun's preflight secret name field is valid", "preflight secret name", preflightSecret.Name)

	// Validate log level
	logLevelLowerCase := strings.ToLower(r.Spec.LogLevel)
	switch logLevelLowerCase {
	case "info", "debug", "warn", "warning", "error":
		cnfcertificationsuiterunlog.Info("CnfCertificationSuiteRun's log level field is valid", "log level", logLevelLowerCase)
	default:
		err = fmt.Errorf("not a valid slog Level: %q", logLevelLowerCase)
		cnfcertificationsuiterunlog.Error(err, "CnfCertificationSuiteRun's log level field is invalid",
			"log level", logLevelLowerCase)
		return err
	}
	cnfcertificationsuiterunlog.Info("test")

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
//
//nolint:revive
func (r *CnfCertificationSuiteRun) ValidateUpdate(old runtime.Object) error {
	cnfcertificationsuiterunlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CnfCertificationSuiteRun) ValidateDelete() error {
	cnfcertificationsuiterunlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
