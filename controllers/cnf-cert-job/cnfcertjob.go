package cnfcertjob

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/greyerof/cnf-certification-operator/controllers/definitions"
)

type Config struct {
	LabelsFilter        string
	LogLevel            string
	ConfigMapName       string
	PreflightSecretName string
}

func NewConfig(labelsFilter, logLevel, configMapName, preflightSecretName string) *Config {
	return &Config{
		LabelsFilter:        labelsFilter,
		LogLevel:            logLevel,
		ConfigMapName:       configMapName,
		PreflightSecretName: preflightSecretName,
	}
}

func New(config *Config, podName string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: "cnf-certification-operator"},
		Spec: corev1.PodSpec{
			ServiceAccountName: "cnf-certification-operator-controller-manager",
			RestartPolicy:      "Never",
			Containers: []corev1.Container{
				{
					Name:  definitions.CnfCertSuiteSidecarContainerName,
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
							Name:  definitions.SideCarResultsFolderEnvVar,
							Value: definitions.CnfCertSuiteResultsFolder,
						},
					},
					ImagePullPolicy: "Always",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "cnf-certsuite-output",
							ReadOnly:  true,
							MountPath: definitions.CnfCertSuiteResultsFolder,
						},
					},
				},
				{
					Name:    definitions.CnfCertSuiteContainerName,
					Image:   "quay.io/testnetworkfunction/cnf-certification-test:unstable",
					Command: []string{"./run-cnf-suites.sh"},
					Args:    []string{"-l", config.LabelsFilter, "-o", definitions.CnfCertSuiteResultsFolder},
					Env: []corev1.EnvVar{
						{
							Name:  "TNF_LOG_LEVEL",
							Value: config.LogLevel,
						},
						{
							Name:  "PFLT_DOCKERCONFIG",
							Value: definitions.PreflightDockerConfigFilePath,
						},
						{
							Name:  "TNF_CONFIGURATION_PATH",
							Value: definitions.CnfCertSuiteConfigFilePath,
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
							MountPath: definitions.CnfCertSuiteResultsFolder,
						},
						{
							Name:      "cnf-certsuite-config",
							ReadOnly:  true,
							MountPath: definitions.CnfCnfCertSuiteConfigFolder,
						},
						{
							Name:      "cnf-certsuite-preflight-dockerconfig",
							ReadOnly:  true,
							MountPath: definitions.CnfPreflightConfigFolder,
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
								Name: config.ConfigMapName,
							},
						},
					},
				},
				{
					Name: "cnf-certsuite-preflight-dockerconfig",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: config.PreflightSecretName,
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{},
	}
}
