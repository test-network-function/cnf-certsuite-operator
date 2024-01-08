package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	"github.com/greyerof/cnf-certification-operator/cnf-cert-sidecar/app/claim"
	cnfcertsuitereport "github.com/greyerof/cnf-certification-operator/cnf-cert-sidecar/app/cnf-cert-suite-report"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sideCarResultsFolderEnvVar = "TNF_RESULTS_FOLDER"
	claimFileName              = "claim.json"
	multiplier                 = 5
)

func handleClaimFile(k8sClient client.Client) {
	claimFolder := os.Getenv(sideCarResultsFolderEnvVar)
	logrus.Infof("Claim file folder: %v", claimFolder)

	claimFilePath := claimFolder + "/" + claimFileName
	for {
		_, err := os.Stat(claimFilePath)
		if os.IsNotExist(err) {
			logrus.Warnf("Claim file not found yet. Waiting 5 secs...")
			time.Sleep(multiplier * time.Second)
			continue
		}

		// Wait extra 5 secs to give time to the writer to finish.
		// TODO: Use file locking mechanism between writer/reader of this file.
		time.Sleep(multiplier * time.Second)

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

		// Create cnfCertSuiteReport
		config := cnfcertsuitereport.NewConfig(&claimContent)
		cnfCertSuiteReport := cnfcertsuitereport.New(config)
		err = k8sClient.Create(context.TODO(), cnfCertSuiteReport)
		if err != nil {
			logrus.Fatalf("Failed to create CnfCertificationSuiteReport object in ns cnf-certification-operator: %v", err)
		}

		cnfcertsuitereport.UpdateStatus(cnfCertSuiteReport, &claimContent.Claim.Results)
		err = k8sClient.Status().Update(context.TODO(), cnfCertSuiteReport)
		if err != nil {
			logrus.Fatalf("Failed to update CnfCertificationSuiteReport.Status object object in ns cnf-certification-operator: %v", err)
		}

		logrus.Infof("CnfCertificationSuiteReport created with results:\n%s", cnfCertSuiteReport.Status.Results)
		break
	}
}

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

	handleClaimFile(k8sClient)
}
