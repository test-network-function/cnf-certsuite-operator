#!/bin/bash
#
# WARNING: This is a helper script to be run with the "timeout" command, like:
# $ timeout 60s $0 pod mypodname mynamespace
#
# Polls the cluster with "oc get ..." every N secs and returns with 0 status
# code as soon as the given resource by name is found in a given namespace.
# In case the resource is not found or there's any problem with "oc get", the
# script will never return, as it's polling every N secs, no matter the type of
# error was returned.
#
# $1 resource kind, e.g. pod, deployment...
# $2 name
# $3 namespace
# $4 check interval time (optional, defaults to 5 segs)
#
# Examples:
#  $0 pod mypodname mypodnamespace

DEFAULT_INTERVAL_CHECK_SEGS=5
INTERVAL_CHECK_SEGS=${DEFAULT_INTERVAL_CHECK_SEGS}

RESOURCE_KIND=$1
RESOURCE_NAME=$2
NAMESPACE=$3

if [ "$4" != "" ] ; then
  INTERVAL_CHECK_SEGS=$4
fi

echo "Polling every ${INTERVAL_CHECK_SEGS} secs for ${RESOURCE_NAME} (kind: ${RESOURCE_KIND}) to exist in namespace ${NAMESPACE}"

while true; do
  if oc get "${RESOURCE_KIND}" -n "${NAMESPACE}" "${RESOURCE_NAME}" ; then
    echo "${RESOURCE_NAME} (kind: ${RESOURCE_KIND}) found in namespace ${NAMESPACE}."
    exit 0
  fi

  echo "${RESOURCE_NAME} (kind: ${RESOURCE_KIND}) not found yet in namespace ${NAMESPACE}..."
  sleep "${INTERVAL_CHECK_SEGS}"
done
