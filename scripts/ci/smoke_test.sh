#!/bin/bash
#
# This is a very simple functional test the controller and the expected
# CnfCertificationSuiteReport CR that should be generated.
#
# Applies the CnfCertificatioSuiteRun sample in config/samples along with its
# related configmap and (preflight) secret. The script waits for the report CR
# (CnfCertificationSuiteReport) to exist in the operator's namespace and then
# checks whether the verdict and the summary counters to have the expected
# values. The expected values can be customized by exporting their env vars
# before running the script.
#
# A patch to the kustomize.yaml file in config/samples needs to be done to add
# the configmap and the secret.
#

# Bash settings: display (expanded) commands and fast exit on first error.
set -o xtrace
set -o errexit

DEFAULT_CNF_CERTSUITE_OPERATOR_NAMESPACE="cnf-certsuite-operator"
CNF_CERTSUITE_OPERATOR_NAMESPACE=${CNF_CERTSUITE_OPERATOR_NAMESPACE:-$DEFAULT_CNF_CERTSUITE_OPERATOR_NAMESPACE}

# Load samples, patching the kustomization.yaml to include the configmap and the preflight dockerconfig.
make deploy-samples

# Apply/create the sample CR.
oc kustomize config/samples | oc apply -f -

# Wait for the CnfCertificationSuiteReport CR to appear in the operator's namespace.
timeout 120s ./scripts/ci/resource_exists.sh cnfcertificationsuitereports cnf-job-run-1-report "${CNF_CERTSUITE_OPERATOR_NAMESPACE}"

# Check that a cnfcertificationsuitereport CR has been created.
# Disable command expansion to avoid output mangling.
set +o xtrace

# Save the report CR in JSON.
reportJson=$(oc get cnfcertificationsuitereports -n "${CNF_CERTSUITE_OPERATOR_NAMESPACE}" cnf-job-run-1-report -o json)

# Show the report CR for debugging purposes.
echo "$reportJson" | jq

# Run checks for verdict and counters.
export EXPECTED_VERDICT=${EXPECTED_VERDICT:-"pass"}
export EXPECTED_TOTAL_TCS=${EXPECTED_TOTAL_TCS:-"88"}
export EXPECTED_FAILED=${EXPECTED_FAILED:-"0"}
export EXPECTED_PASSED=${EXPECTED_PASSED:-"4"}
export EXPECTED_SKIPPED=${EXPECTED_SKIPPED:-"84"}

# Check the verdit is pass
echo "$reportJson" | jq 'if .status.verdict == env.EXPECTED_VERDICT then "verdict is "+env.EXPECTED_VERDICT else error("verdict mismatch: \(.status.verdict), expected "+env.EXPECTED_VERDICT) end'
echo "$reportJson" | jq 'if .status.summary.total   | tostring == env.EXPECTED_TOTAL_TCS then "total tc num is ok"   else error("total tcs mismatch: \(.status.summary.total), expected "+env.EXPECTED_TOTAL_TCS) end'
echo "$reportJson" | jq 'if .status.summary.passed  | tostring == env.EXPECTED_PASSED    then "passed tc num is ok"  else error("passed tcs mismatch: \(.status.summary.passed), expected "+env.EXPECTED_PASSED) end'
echo "$reportJson" | jq 'if .status.summary.skipped | tostring == env.EXPECTED_SKIPPED   then "skipped tc num is ok" else error("skipped tcs mismatch: \(.status.summary.skipped), expected "+env.EXPECTED_SKIPPED) end'
echo "$reportJson" | jq 'if .status.summary.failed  | tostring == env.EXPECTED_FAILED    then "failed tc num is ok" else error("failed tcs mismatch: \(.status.summary.failed), expected "+env.EXPECTED_FAILED) end'
