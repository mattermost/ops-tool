#!/bin/bash

kubectx ${K8S_CLUSTER} &> /dev/null
    
LOG=$(kubectl -n ${K8S_NAMESPACE} logs --tail=500 deployment.apps/${DEPLOYMENT_NAME})

jq --null-input \
   --arg log "$LOG" \
   '{status:"ok", "data": $log }'
