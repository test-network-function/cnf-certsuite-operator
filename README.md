# CNF Certification Suite Operator

[![red hat](https://img.shields.io/badge/red%20hat---?color=gray&logo=redhat&logoColor=red&style=flat)](https://www.redhat.com)
[![openshift](https://img.shields.io/badge/openshift---?color=gray&logo=redhatopenshift&logoColor=red&style=flat)](https://www.redhat.com/en/technologies/cloud-computing/openshift)

## Description

Kubernetes/Openshift Operator (scaffolded with operator-sdk) running the
[CNF Certification Suite Container](https://github.com/test-network-function/cnf-certification-test).

The CNF Certification Suites provide a set of test cases for the
Containerized Network Functions/Cloud Native Functions (CNFs) to verify if
best practices for deployment on Red Hat OpenShift clusters are followed.

### How does it work?

The Operator uses a CR representing a CNF Certification Suites run.
In order to run the suites, such "run" CR has to be created together
with a Config Map containing the cnf certification suites configuration,
and a Secret containing the preflight suite credentials.

See resources relationship diagram:

![run config](doc/uml/run_config.png)

When the CR is deployed, a new pod with two containers is created:

1. Container built with the cnf certification image in order to run the suites.
2. Container which creates a new CR representing the CNF Certification suites
results based on results claim file created by the previous container.

    See container's flow in the following diagram:

    ![side car](doc/uml/side_car.png)

**See diagram summarizing the process:**

![Use Case Run](doc/uml/use_case_run.png)

## Getting Started

You’ll need a Kubernetes cluster to run against.
You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing,
or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your
kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Install operator

#### Iniital steps

1. Clone Cnf Certification Operator repo:

    ```sh
    git clone https://github.com/greyerof/tnf-op.git
    ```

2. Install cert-manager:

    ```sh
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
    ```

#### Option 1: Use a your own registry account

1. Export images environment variables:

    ```sh
    export IMG=<your-registry.com>/<your-repo>/cnf-certsuite-operator:<version>
    export SIDECAR_IMG=<your-registry.com>/<your-repo>/cnf-certsuite-operator-sidecar:<version>
    ```

2. Build and upload the controller image to your registry account:

    ```sh
    make docker-build docker-push
    ```

3. Build and upload the side car image to your registry account:

    ```sh
    docker build -f cnf-cert-sidecar/Dockerfile -t $SIDECAR_IMG .
    docker push $SIDECAR_IMG
    ```

4. Deploy the operator, using the previously uploaded controller image,
 and the built side car image:

    ```sh
    make deploy
    ```

#### Option 2: Use local images

1. Export images environment variables:

    ```sh
    export IMG=<your-cnf-certsuite-operator-image-name>
    export SIDECAR_IMG=<your-sidecar-app-image-name>
    ```

2. Build controller and side car images:

    ```sh
    scripts/ci/build.sh
    ```

3. Deploy previously built images by preloading them into the kind cluster's nodes:

    ```sh
    scripts/ci/deploy.sh
    ```

### Test it out

Use our samples to test out the cnf certification operator, with the following command:

```sh
make deploy-samples
```

### Running test suites on the cluster

1. Create Resources

    In order to use the cnf certification suite operator,
    you'll have to create yaml files for the following resources:

    1. Config map:\
    Containing the cnf certification configuration file
    content under the `tnf_config.yaml` key.\
    (see [CNF Certification configuration description](https://test-network-function.github.io/cnf-certification-test/configuration/))

    2. Secret:\
    Containing cnf preflight suite credentials
    under the `preflight_dockerconfig.json` key.\
    (see [Preflight Integration description](https://test-network-function.github.io/cnf-certification-test/runtime-env/#disable-intrusive-tests))

    3. CnfCertificationSuiteRun CR:\
    Containing the following Spec fields that has to be filled in:
        - **labelsFilter**: Wanted label filtering the cnf certification tests suite.
        - **logLevel**: Wanted log level of cnf certification tests suite run.\
        Log level options: "info", "debug", "warn", "error"
        - **timeout**: Wanted timeout for the the cnf certification tests.
        - **configMapName**: Name of the config map defined at stage 1.
        - **preflightSecretName**: Name of the preflight Secret
        defined at stage 2.
        - **enableDataCollection**: Set to "true" to enable data collection,
        or "false" otherwise\
        **Note:** Current operator's version **doesn't** support
        setting enableDataCollection to "true".

        See a [sample CnfCertificationSuiteRun CR](https://github.com/greyerof/tnf-op/blob/main/config/samples/cnf-certifications_v1alpha1_cnfcertificationsuiterun.yaml)

    **Note:** All resources have to be defined
    under the `cnf-certsuite-operator` namespace.

2. Apply resources into the cluster

    After creating all the yaml files for required resources,
    use the following commands to apply them into the cluster:

    ```sh
    oc apply -f /path/to/config/map.yaml
    oc apply -f /path/to/preflight/secret.yaml
    oc apply -f /path/to/cnfCertificationSuiteRun.yaml
    ```

    **Note**: The same config map and secret can be reused
    by different CnfCertificationSuiteRun CR's.

### Review results

If all of the resources were applied successfully, the cnf certification suites
will run on a new created `pod` in the `cnf-certsuite-operator` namespace.

Check whether the pod creation and the cnf certification suites run were successful
by checking CnfCertificationSuiteRun CR's status, using the following command:

```sh
oc get cnfcertificationsuiteruns.cnf-certifications.redhat.com -n cnf-certsuite-operator
```

In the successful case, expect to see the following status:

```sh
NAME                              AGE   STATUS
cnfcertificationsuiterun-sample   50m   CertSuiteFinished
```

When the pod is completed, a new `CnfCertificationSuiteReport` will be created
under the same namespace.
CNF certification suites results will be stored in the CR's Status different fields:

- Results: For every test case, contains its result and logs.
If the the result is "skipped" or "failed" contains also the skip\failure reason.

    See example:

    ```sh
    Logs:            INFO  [Feb 15 10:46:43.050] [check.go: 263] [observability-crd-status]
    Running check (labels: [common observability-crd-status observability]) 
    INFO  [Feb 15 10:46:43.050] [suite.go: 144] [observability-crd-status] 
    Testing CRD: crdexamples.test-network-function.com
    INFO  [Feb 15 10:46:43.050] [suite.go: 153] [observability-crd-status] 
    CRD: crdexamples.test-network-function.com, version: v1 has a status subresource
    INFO  [Feb 15 10:46:43.050] [checksdb.go: 115] [observability-crd-status] 
    Recording result "PASSED", 
    claimID: {Id:observability-crd-status Suite:observability Tags:common}

    Result:          passed
    Test Case Name:  observability-crd-status
    Logs: INFO  [Feb 15 10:46:43.025] [checksgroup.go: 83]
    [access-control-net-raw-capability-check] 
    Skipping check access-control-net-raw-capability-check, reason: no matching labels
    INFO  [Feb 15 10:46:43.026] [checksdb.go: 115]
    [access-control-net-raw-capability-check] Recording result "SKIPPED",
    claimID: {Id:access-control-net-raw-capability-check Suite:access-control Tags:telco}

    Reason:          no matching labels
    Result:          skipped
    Test Case Name:  access-control-net-raw-capability-check
    Logs:            INFO  [Feb 15 10:46:43.025] [checksgroup.go: 83]
    [access-control-security-context-non-root-user-check] Skipping
    checkaccess-control-security-context-non-root-user-check,
    reason: no matching labels
    INFO  [Feb 15 10:46:43.026] [checksdb.go: 115]
    [access-control-security-context-non-root-user-check] Recording result "SKIPPED",
    claimID: {Id:access-control-security-context-non-root-user-check Suite
    :access-control Tags:common}

    Reason:          no matching labels
    Result:          skipped
    Test Case Name:  access-control-security-context-non-root-user-check
    ```

- Summary: Summarize the total number of tests by their results.
- Verdict: Specifies the overall result of the CNF certificattion suites run.\
Poissible verdicts: "pass", "skip", "fail", "error".

Run the following command to ensure its creation:

```sh
oc get cnfcertificationsuitereports.cnf-certifications.redhat.com -n cnf-certsuite-operator
```

To review the test results describe the created
`CnfCertificationSuiteReport` run the following command:

```sh
oc describe cnfcertificationsuitereports.cnf-certifications.redhat.com \
-n cnf-certsuite-operator <report`s name>
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

```plaintext
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
