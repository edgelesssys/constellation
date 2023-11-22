#!/usr/bin/env bash

set -euo pipefail

function debug() {
  echo "DEBUG: $*" >&2
}

function warn() {
  echo "WARN: $*" >&2
}

function inputs() {
  name="${1}"
  local val
  val=$(jq -r ".\"${name}\"" "${inputFile}")
  if [[ ${val} == "null" ]]; then
    warn "Input ${name} not found in ${inputFile}"
    return
  fi
  echo "${val}"
}

function flagsFromInput() {
  flagNames=("${@}")
  for name in "${flagNames[@]}"; do
    val=$(inputs "${name}")
    if [[ -n ${val} ]]; then
      echo "--${name}=${val}"
    fi
  done
}

function createIssue() {
  flags=(
    "assignee"
    "body"
    "body-file"
    "label"
    "milestone"
    "project"
    "template"
    "title"
  )
  readarray -t flags <<< "$(flagsFromInput "${flags[@]}")"
  flags+=("--repo=$(inputs owner)/$(inputs repo)")
  debug gh issue create "${flags[@]}"
  gh issue create "${flags[@]}"
}

function listProjects() {
  flags=(
    "owner"
  )
  readarray -t flags <<< "$(flagsFromInput "${flags[@]}")"
  flags+=("--format=json")
  debug gh project list "${flags[@]}"
  gh project list "${flags[@]}" >> projects.json
}

function findProjectID() {
  project=$(inputs "project")
  out="$(
    jq -r \
      --arg project "${project}" \
      '.projects[]
        | select(.title == $project)
        | .id' \
      projects.json
  )"
  debug "Project ID: ${out}"
  echo "${out}"
}

function findProjectNo() {
  project=$(inputs "project")
  out="$(
    jq -r \
      --arg project "${project}" \
      '.projects[]
        | select(.title == $project)
        | .number' \
      projects.json
  )"
  debug "Project Number: ${out}"
  echo "${out}"
}

function listItems() {
  local projectNo="${1}"
  flags=(
    "owner"
  )
  readarray -t flags <<< "$(flagsFromInput "${flags[@]}")"
  flags+=("--limit=1000")
  flags+=("--format=json")
  debug gh project item-list "${flags[@]}" "${projectNo}"
  gh project item-list "${flags[@]}" "${projectNo}" >> issues.json
}

function findIssueItemID() {
  local issueURL="${1}"
  out="$(
    jq -r \
      --arg issueURL "${issueURL}" \
      '.items[]
        | select(.content.url == $issueURL)
        | .id' \
      issues.json
  )"
  debug "Issue Item ID: ${out}"
  echo "${out}"
}

function listFields() {
  local projectNo="${1}"
  flags=(
    "owner"
  )
  readarray -t flags <<< "$(flagsFromInput "${flags[@]}")"
  flags+=("--limit=1000")
  flags+=("--format=json")
  debug gh project field-list "${flags[@]}" "${projectNo}"
  gh project field-list "${flags[@]}" "${projectNo}" >> fields.json
}

function findFieldID() {
  local fieldName="${1}"
  out="$(
    jq -r \
      --arg fieldName "${fieldName}" \
      '.fields[]
        | select(.name == $fieldName)
        | .id' \
      fields.json
  )"
  debug "Field ID of '${fieldName}': ${out}"
  echo "${out}"
}

function findSelectFieldID() {
  local fieldName="${1}"
  local fieldValue="${2}"
  out="$(
    jq -r \
      --arg fieldName "${fieldName}" \
      --arg fieldValue "${fieldValue}" \
      '.fields[]
        | select(.name == $fieldName)
        | .options[]
        | select(.name == $fieldValue)
        | .id' \
      fields.json
  )"
  debug "Field ID of '${fieldName}': ${out}"
  echo "${out}"
}

function findFieldType() {
  local fieldName="${1}"
  out="$(
    jq -r \
      --arg fieldName "${fieldName}" \
      '.fields[]
        | select(.name == $fieldName)
        | .type' \
      fields.json
  )"
  debug "Field type of '${fieldName}': ${out}"
  echo "${out}"
}

function editItem() {
  local projectID="${1}"
  local itemID="${2}"
  local id="${3}"
  local value="${4}"
  flags=(
    "--project-id=${projectID}"
    "--id=${itemID}"
    "--field-id=${id}"
    "--text=${value}"
  )
  debug gh project item-edit "${flags[@]}"
  gh project item-edit "${flags[@]}" > /dev/null
}

function setFields() {
  local projectID="${1}"
  local itemID="${2}"

  fieldsLen="$(jq -r '.fields' "${inputFile}" | yq 'length')"
  debug "Number of fields in input: ${fieldsLen}"
  for ((i = 0; i < fieldsLen; i++)); do
    name="$(jq -r '.fields' "${inputFile}" |
      yq "to_entries | .[${i}].key")"
    value="$(jq -r '.fields' "${inputFile}" |
      yq "to_entries | .[${i}].value")"
    debug "Field ${i}: ${name} = ${value}"
    type=$(findFieldType "${name}")

    case "${type}" in
    "ProjectV2Field")
      id=$(findFieldID "${name}")
      ;;
    "ProjectV2SingleSelectField")
      id=$(findSelectFieldID "${name}" "${value}")
      ;;
    *)
      warn "Unknown field type: ${type}"
      return 1
      ;;
    esac

    editItem "${projectID}" "${itemID}" "${id}" "${value}"
  done
}

function main() {
  inputFile="$(realpath "${1}")"

  workdir=$(mktemp -d)
  pushd "${workdir}" > /dev/null
  trap 'debug "not cleaning up, working directory at: ${workdir}"' ERR

  issueURL=$(createIssue)
  echo "${issueURL}"

  project=$(inputs "project")
  if [[ -z ${project} ]]; then
    return
  fi

  listProjects
  projectNo=$(findProjectNo)
  projectID=$(findProjectID)

  listItems "${projectNo}"
  issueItemID=$(findIssueItemID "${issueURL}")
  listFields "${projectNo}"

  setFields "${projectID}" "${issueItemID}"

  popd > /dev/null
  rm -rf "${workdir}"
}

main "${@}"
