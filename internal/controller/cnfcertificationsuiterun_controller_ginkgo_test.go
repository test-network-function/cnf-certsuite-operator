/*
Copyright 2024.

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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cnfcertificationsv1alpha1 "github.com/test-network-function/cnf-certsuite-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var runCR *cnfcertificationsv1alpha1.CnfCertificationSuiteRun

var _ = Describe("CnfCertificationSuiteRun Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "cnfcertificationsuiterun-sample"

		ctx := context.Background()

		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cnf-certsuite-operator",
			},
		}
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "cnf-certsuite-operator",
		}
		runCR = &cnfcertificationsv1alpha1.CnfCertificationSuiteRun{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind CnfCertificationSuiteRun")
			err := k8sClient.Get(ctx, typeNamespacedName, runCR)
			if err != nil && errors.IsNotFound(err) {
				resource := &cnfcertificationsv1alpha1.CnfCertificationSuiteRun{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "cnf-certsuite-operator",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &cnfcertificationsv1alpha1.CnfCertificationSuiteRun{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance CnfCertificationSuiteRun")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CnfCertificationSuiteRunReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
