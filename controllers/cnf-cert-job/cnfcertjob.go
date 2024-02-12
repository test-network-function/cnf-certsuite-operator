package cnfcertjob

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/greyerof/cnf-certification-operator/controllers/definitions"
)

const (
	// Be careful when changing this SA name.
	// 1. It must match the flag --extra-service-accounts in "make bundle".
	// 2. The prefix is "cnf-certsuite-". It should match the field namePrefix field in config/default/kustomization.yaml.
	clusterAccessServiceAccountName = "cnf-certsuite-cluster-access"
)

func New(options ...func(*corev1.Pod)) *corev1.Pod {
	jobPod := newInitialJobPod()

	for _, o := range options {
		o(jobPod)
	}
	return jobPod
}

//nolint:funlen
func newInitialJobPod() *corev1.Pod {
	return &corev1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: corev1.PodSpec{
			ServiceAccountName: clusterAccessServiceAccountName,
			RestartPolicy:      "Never",
			Containers: []corev1.Container{
				{
					Name: definitions.CnfCertSuiteSidecarContainerName,
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
							Name: "MY_POD_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
						{
							Name:  definitions.SideCarResultsFolderEnvVar,
							Value: definitions.CnfCertSuiteResultsFolder,
						},
					},
					ImagePullPolicy: "IfNotPresent",
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
					Args:    []string{"-o", definitions.CnfCertSuiteResultsFolder},
					Env: []corev1.EnvVar{
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
			},
		},
		Status: corev1.PodStatus{},
	}
}

func WithPodName(podName string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		p.ObjectMeta.Name = podName
	}
}

func WithNamespace(namespace string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		p.ObjectMeta.Namespace = namespace
	}
}

func WithCertSuiteConfigRunName(certSuiteConfigRunName string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		envVar := corev1.EnvVar{Name: "RUN_CR_NAME", Value: certSuiteConfigRunName}
		p.Spec.Containers[0].Env = append(p.Spec.Containers[0].Env, envVar)
	}
}

func WithLabelsFilter(labelsFilter string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		p.Spec.Containers[1].Args = append(p.Spec.Containers[1].Args, "-l", labelsFilter)
	}
}

func WithLogLevel(loglevel string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		envVar := corev1.EnvVar{Name: "TNF_LOG_LEVEL", Value: loglevel}
		p.Spec.Containers[1].Env = append(p.Spec.Containers[1].Env, envVar)
	}
}

func WithTimeOut(timeout string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		envVar := corev1.EnvVar{Name: "TIMEOUT", Value: timeout}
		p.Spec.Containers[1].Env = append(p.Spec.Containers[1].Env, envVar)
	}
}

func WithConfigMap(configMapName string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		Volume := corev1.Volume{
			Name: "cnf-certsuite-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, Volume)
	}
}

func WithPreflightSecret(preflightSecretName string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		Volume := corev1.Volume{
			Name: "cnf-certsuite-preflight-dockerconfig",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: preflightSecretName,
				},
			},
		}
		p.Spec.Volumes = append(p.Spec.Volumes, Volume)
	}
}

func WithSideCarApp(sideCarAppImage string) func(*corev1.Pod) {
	return func(p *corev1.Pod) {
		p.Spec.Containers[0].Image = sideCarAppImage
	}
}
