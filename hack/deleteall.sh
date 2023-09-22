deleteAll() {
    res=$1
    for name in `oc get $res -n cnf-certification-operator -o json | jq -r '.items[] | .metadata.name' | sort -r` ; do oc delete $res -n cnf-certification-operator $name ; done
}

deleteAll cnfcertificationsuiteruns
deleteAll cnfcertificationsuitereports
deleteAll pods

oc get all -n cnf-certification-operator
