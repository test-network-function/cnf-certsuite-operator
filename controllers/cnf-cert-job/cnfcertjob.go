package cnfcertjob

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cnfcertificationsv1alpha1 "github.com/greyerof/cnf-certification-operator/api/v1alpha1"
	"github.com/greyerof/cnf-certification-operator/controllers/controllerhelper"
)

func NewConfig(cnfrun cnfcertificationsv1alpha1.CnfCertificationSuiteRun, cnfRunPodId int) corev1.Pod {
	cnfCertJobPod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", controllerhelper.CnfCertPodNamePrefix, cnfRunPodId),
			Namespace: "cnf-certification-operator"},
		Spec: corev1.PodSpec{
			ServiceAccountName: "cnf-certification-operator-controller-manager",
			RestartPolicy:      "Never",
			Containers: []corev1.Container{
				{
					Name:  "cnf-certsuite-sidecar",
					Image: "quay.io/greyerof/cnf-op:sidecarv2",
					Env: []corev1.EnvVar{
						{
							Name: "MY_POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name:  controllerhelper.SideCarResultsFolderEnvVar,
							Value: controllerhelper.CnfCertSuiteResultsFolder,
						},
					},
					ImagePullPolicy: "Always",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "cnf-certsuite-output",
							ReadOnly:  true,
							MountPath: controllerhelper.CnfCertSuiteResultsFolder,
						},
					},
				},
				{
					Name:    "cnf-certsuite",
					Image:   "quay.io/testnetworkfunction/cnf-certification-test:unstable",
					Command: []string{"./run-cnf-suites.sh"},
					Args:    []string{"-l", cnfrun.Spec.LabelsFilter, "-o", controllerhelper.CnfCertSuiteResultsFolder},
					Env: []corev1.EnvVar{
						{
							Name:  "TNF_LOG_LEVEL",
							Value: cnfrun.Spec.LogLevel,
						},
						{
							Name:  "PFLT_DOCKERCONFIG",
							Value: controllerhelper.PreflightDockerConfigFilePath,
						},
						{
							Name:  "TNF_CONFIGURATION_PATH",
							Value: controllerhelper.CnfCertSuiteConfigFilePath,
						},
						{
							Name:  "TNF_NON_INTRUSIVE_ONLY",
							Value: "true",
						},
					},
					ImagePullPolicy: "Always",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "cnf-certsuite-output",
							MountPath: controllerhelper.CnfCertSuiteResultsFolder,
						},
						{
							Name:      "cnf-certsuite-config",
							ReadOnly:  true,
							MountPath: controllerhelper.CnfCnfCertSuiteConfigFolder,
						},
						{
							Name:      "cnf-certsuite-preflight-dockerconfig",
							ReadOnly:  true,
							MountPath: controllerhelper.CnfPreflightConfigFolder,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "cnf-certsuite-output",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "cnf-certsuite-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: cnfrun.Spec.ConfigMapName,
							},
						},
					},
				},
				{
					Name: "cnf-certsuite-preflight-dockerconfig",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: cnfrun.Spec.PreflightSecretName,
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{},
	}
	return cnfCertJobPod
}

func Deploy(cnfCertJobPod corev1.Pod, r *controllerhelper.CnfCertificationSuiteRunReconciler, ctx context.Context) {
	err := r.Create(ctx, &cnfCertJobPod)
	if err != nil {
		log.Log.Error(err, "Failed to create CNF Cert job")
	}
}
