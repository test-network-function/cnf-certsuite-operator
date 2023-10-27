package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	"github.com/greyerof/cnf-certification-operator/cnf-cert-sidecar/app/claim"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sideCarResultsFolderEnvVar = "TNF_RESULTS_FOLDER"
	claimFileName              = "claim.json"
)

// This CNF Certification sidecar container expects to be running in the same
// pod as the CNF Cert Suite container.
//
// The sidecar will wait for TNF_CNF_CERT_TIMEOUT secs until the claim.json output is present in a
// shared volume folder. Then it will parse it to get the list of test cases
// results, and creates the CR CnfCertificationSuiteReport
func main() {
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		logrus.Fatalf("failed to add to scheme: %v", err)
	}

	err = cnfcertificationsv1alpha1.AddToScheme(scheme)
	if err != nil {
		logrus.Fatalf("failed to add cnfcertificationsv1alpha1 to scheme: %v", err)
	}

	// k8sRestConfig, err := rest.InClusterConfig()
	kubeConfigFile := os.Getenv("HOME") + "/.kube/config"
	_, err = os.Stat(kubeConfigFile)
	if err != nil {
		kubeConfigFile = ""
	}

	k8sRestConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		logrus.Fatalf("Failed to get incluster rest config: %v", err)
	}

	DefaultTimeout := 10 * time.Second
	k8sRestConfig.Timeout = DefaultTimeout

	k8sClient, err := client.New(k8sRestConfig, client.Options{Scheme: scheme})
	if err != nil {
		logrus.Fatalf("Failed to get k8s client: %v", err)
	}

	claimFolder := os.Getenv(sideCarResultsFolderEnvVar)
	logrus.Infof("Claim file folder: %v", claimFolder)

	claimFilePath := claimFolder + "/" + claimFileName
	for {
		_, err := os.Stat(claimFilePath)
		if os.IsNotExist(err) {
			logrus.Warnf("Claim file not found yet. Waiting 5 secs...")
			time.Sleep(5 * time.Second)
			continue
		}

		// Wait extra 5 secs to give time to the writer to finish.
		// TODO: Use file locking mechanism between writer/reader of this file.
		time.Sleep(5 * time.Second)

		logrus.Infof("Claim file found at %v", claimFilePath)
		claimBytes, err := os.ReadFile(claimFilePath)
		if err != nil {
			logrus.Fatalf("Failed to read claim file %s: %v", claimFilePath, err)
		}

		claimContent := claim.Schema{}
		err = json.Unmarshal(claimBytes, &claimContent)
		if err != nil {
			logrus.Fatalf("Failed to unmarshal claim json: %v", err)
		}

		results := []cnfcertificationsv1alpha1.TestCaseResult{}
		for i := range claimContent.Claim.Results {
			tsResults := claimContent.Claim.Results[i]
			for j := range tsResults {
				tcResult := tsResults[j]
				results = append(results, cnfcertificationsv1alpha1.TestCaseResult{
					TestCaseName: tcResult.TestID.ID,
					Result:       tcResult.State,
				})
			}
		}

		reportCrName := fmt.Sprintf("%s-report", os.Getenv("MY_POD_NAME"))
		err = k8sClient.Create(context.TODO(), &cnfcertificationsv1alpha1.CnfCertificationSuiteReport{
			ObjectMeta: metav1.ObjectMeta{Name: reportCrName, Namespace: "cnf-certification-operator"},
			Spec:       cnfcertificationsv1alpha1.CnfCertificationSuiteReportSpec{Results: results},
			Status:     cnfcertificationsv1alpha1.CnfCertificationSuiteReportStatus{},
		})
		if err != nil {
			logrus.Fatalf("Failed to create CnfCertificationSuiteReport object in ns cnf-certification-operator: %v", err)
		}

		logrus.Infof("CnfCertificationSuiteReport created with results:\n%s", results)
		break
		// logrus.Infof("CnfCertificationSuiteReport created succesfully. Waiting 60 secs because why not...")

		// time.Sleep(60 * time.Second)
	}
}
