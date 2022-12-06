#!/bin/sh

function get_kyma_status {
	local number=1
	while [[ $number -le 12*4 ]] ; do
		echo ">--> checking kyma status #$number"
		local STATUS=$(kubectl get eventing -n kcp-system default-kyma-eventing -o jsonpath='{.status.state}')
		echo "kyma status: ${STATUS:='UNKNOWN'}"
		[[ "$STATUS" == "Ready" ]] && return 0
		sleep 5
        	((number = number + 1))
	done

	kubectl get all --all-namespaces
	exit 1
}

get_kyma_status
