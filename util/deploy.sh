#!/bin/bash

# Helper script for using a standardized version flag when deploying.

REPO_DIR="$(dirname "${BASH_SOURCE[0]}")/.."
source "${REPO_DIR}/util/logging.sh"
source "${REPO_DIR}/util/path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${REPO_DIR})}

usage() {
  USAGE="Usage: deploy.sh [-p] [-q] [-b] [-h] [app path]
    -p : Production deploy
    -q : Quiet (no user prompts, debugging off)
    -b : Branch name - defaults to current Git branch
    -h : Show (this) help information
    app path: wpt.fyi relative path for the app, e.g. \"webapp\""
  echo "${USAGE}"
}

APP_PATH=${@: -1}
while getopts ':b:phq:g:' flag; do
  case "${flag}" in
    b) BRANCH_NAME="${OPTARG}" ;;
    p) PRODUCTION='true' ;;
    q) QUIET='true' ;;
    h|*) usage && exit 0;;
  esac
done

if [[ "${APP_PATH}" == ""  ]]; then fatal "app path not specified."; fi
if [[ "${APP_PATH}" != "webapp" && "${APP_PATH}" != "results-processor" ]];
then
  fatal "Unrecognized app path \"${APP_PATH}\"."
fi

# Ensure dependencies are installed.
if [[ -z "${QUIET}" ]]; then info "Installing dependencies..."; fi
cd ${WPTD_PATH}
make webserver_deps || fatal "Error installing deps"

# Create a name for this version
BRANCH_NAME=${BRANCH_NAME:-"$(git rev-parse --abbrev-ref HEAD)"}
USER="$(git remote -v get-url origin | sed -E 's#(https?:\/\/|git@)github.com(\/|:)##' | sed 's#/.*$##')-"
if [[ "${USER}" == "web-platform-tests-" ]]; then USER=""; fi

VERSION="${USER}${BRANCH_NAME}"
PROMOTE="--no-promote"

if [[ -n ${PRODUCTION} ]]
then
  if [[ -z "${QUIET}" ]]; then debug "Producing production configuration..."; fi
  if [[ "${USER}" != "" ]]
  then
    if [[ -z "${QUIET}" ]]
    then
      confirm "Are you sure you want to be deploying a non-web-platform-tests repo (${USER})?"
      if [ "${?}" != "0" ]; then fatal "User cancelled the deploy"; fi
    fi
  fi
  # Use SHA for prod-pushes.
  VERSION="$(git rev-parse --short HEAD)"
  PROMOTE="--promote"
fi

if [[ -n "${QUIET}" ]]
then
    QUIET_FLAG="-q"
else
    QUIET_FLAG=""
fi
COMMAND="gcloud app deploy ${PROMOTE} ${QUIET_FLAG} --version=${VERSION} ${APP_PATH}"

if [[ -z "${QUIET}" ]]
then
    info "Deploy command:\n${COMMAND}"
    confirm "Execute?"
    if [[ "${?}" != "0" ]]; then fatal "User cancelled the deploy"; fi
fi

set -e

if [[ -z "${QUIET}" ]]; then info "Executing..."; fi
${COMMAND} || fatal "Deploy returned non-zero exit code $?"

# Comment on the PR if running from Travis.
DEPLOYED_URL=$(gcloud app versions describe ${VERSION} -s default | grep -Po 'versionUrl: \K.*$')
echo "Deployed to ${DEPLOYED_URL}"

exit 0
