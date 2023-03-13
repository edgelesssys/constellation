#!/usr/bin/env bash

set -euo pipefail
set -o errtrace
shopt -s inherit_errexit

function printErr {
  echo -e "\033[0;31mERROR:\033[0m ${1}"
}

function _exitHandler {
  local exit_code=$1
  if [[ ${exit_code} -ne 0 ]]; then
    printErr "$0: exit status ${exit_code}"
  fi
}

function _errorHandler {
  local line=$1
  local linecallfunc=$2
  local command="$3"
  local funcstack="$4"
  printErr "$0: '${command}' failed at line ${line}"

  if [[ ${funcstack} != "::" ]]; then
    echo -ne "\tin ${funcstack} "
    if [[ ${linecallfunc} != "" ]]; then
      echo "called at line ${linecallfunc}"
    else
      echo
    fi
  fi

}

_exitHandlers=()
_errorHandlers=()

function registerExitHandler {
  _exitHandlers+=("$1")
}

function registerErrorHandler {
  _errorHandler+=("$1")
}

function _callExitHandlers {
  _exitHandlers+=(_exitHandler) # Add our handler last.
  for h in "${_exitHandlers[@]}"; do
    ${h} "$@"
  done
}

function _callErrorHandlers {
  _errorHandlers+=(_errorHandler) # Add our handler last.
  for h in "${_errorHandlers[@]}"; do
    ${h} "$@"
  done
}

trap '_callErrorHandlers $LINENO $BASH_LINENO "$BASH_COMMAND" $(printf "::%s" ${FUNCNAME[@]:-})' ERR
trap '_callExitHandlers $?' EXIT
