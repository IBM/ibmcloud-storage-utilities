#!/bin/bash
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Kubernetes Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2022 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

if [ "$TRAVIS_PULL_REQUEST" != "false" ] && [ "$TRAVIS_GO_VERSION" == "tip" ]; then
	curl -s -k -X GET -H "Content-Type: application/json" -H "Accept: application/vnd.travis-ci.2+json"  -H "Authorization: token $TRAVIS_TOKEN"  https://travis-ci.com/github/IBM/ibmcloud-storage-utilities/builds/"$TRAVIS_BUILD_ID" | jq '.jobs[0].state' | sed 's/"//g'> state.out
	RESULT=$(<state.out)
	if [ "$RESULT" != "failed" ]; then
		RESULT_MESSAGE=":warning: Build failed with **tip** version."
		curl -X POST -H "Authorization: token $GHE_TOKEN" https://github.com/IBM/ibmcloud-storage-utilities/repos/"$TRAVIS_REPO_SLUG"/issues/"$TRAVIS_PULL_REQUEST"/comments -H 'Content-Type: application/json' --data '{"body": "'"$RESULT_MESSAGE"'"}'
	fi
fi
