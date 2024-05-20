#!/bin/bash --posix

# -*-  Coding: UTF-8  -*- #
# -*-  System: Linux  -*- #
# -*-  Usage:   *.*   -*- #

# Author: Jacob B. Sanders

# initialize.bash
# - Establishes the base directory and executes the kubebuilder init command
# - https://book.kubebuilder.io/quick-start

# --------------------------------------------------------------------------------
# Bash Set-Options Reference
# --------------------------------------------------------------------------------

# 0. An Opinionated, Well Agreed Upon Standard for Bash Script Execution
# 1. set -o verbose     ::: Print Shell Input upon Read
# 2. set -o allexport   ::: Export all Variable(s) + Function(s) to Environment
# 3. set -o errexit     ::: Exit Immediately upon Pipeline'd Failure
# 4. set -o monitor     ::: Output Process-Separated Command(s)
# 5. set -o privileged  ::: Ignore Externals - Ensures of Pristine Run Environment
# 6. set -o xtrace      ::: Print a Trace of Simple Commands
# 7. set -o braceexpand ::: Enable Brace Expansion
# 8. set -o no-exec     ::: Bash Syntax Debugging
# 9. set -o pipefail    ::: All pipe'd (|) Returns Must Succeed

set -o errexit      # (3)
set -o pipefail     # (9)

# --name "test" --domain "${NAME}.operators.ethr.gg" --repo ${DOMAIN}/${NAME}
function main() {
    if (( ${#@} == 0 )); then
        echo "Argument(s) Required: --name \"<directory-name>\""
        exit 1
    fi

    while [[ $# -gt 0 ]]; do
        case ${1} in
            -n|--name)
                local declare NAME="$2"
                shift # past argument
                shift # past value
                ;;
            -o|--owner)
                local declare OWNER="$2"
                shift # past argument
                shift # past value
                ;;
            -*|--*)
                echo "Unknown: {$1}"
                exit 1
                ;;
            *)
                echo "Positional Arguments Not Allowed"
                exit 1
                ;;
        esac
    done

    if [[ "${NAME}" == "" ]]; then
        echo "The --name flag must be provided a valid string value."
        exit 1
    fi

    if [[ "${OWNER}" == "" ]]; then
        echo "The --owner flag must be provided a valid string value."
        exit 1
    fi

    local declare CWD="$(pwd)"
    if [[ "$(basename "$(git rev-parse --show-toplevel)")" == "local-kind-cluster" ]]; then
        cd "$(git rev-parse --show-toplevel)/operators"
    fi

    local declare DOMAIN="${NAME}.operators.ethr.gg"

    echo " - Domain: ${DOMAIN}"
    echo " - Directory: ${NAME}"
    echo " - Command: kubebuilder init --plugins \"go/v4\" --domain \"${DOMAIN}\" --repo \"${DOMAIN}/${NAME}\" --license \"none\" --owner \"${OWNER}\""
    echo "kubebuilder init --plugins \"go/v4\" --domain \"${DOMAIN}\" --repo \"${DOMAIN}/${NAME}\" --license \"none\" --owner \"${OWNER}\"" > ".command.${NAME}"

    mkdir -p "${NAME}" && cd "${NAME}"

    kubebuilder init --domain "${DOMAIN}" --repo "${DOMAIN}/${NAME}" --license "none"

    echo "0.0.1" > VERSION

    go mod vendor && cd "${CWD}"
}

main "${@}"
