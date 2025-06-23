#!/bin/bash

LOCAL_HOST="your_host"

MAIN_GERRIT="your_site"
STRICT_GERRIT="your_sit"

JARVIS_VERSION_SHEET="true"
IS_TEMP_SHEET="false"

check_environment() {
    local main_gerrit="${MAIN_GERRIT}"
    local local_gerrit=""

    if [[ "${main_gerrit}" != "10.67.16.29" ]]; then
        local_gerrit="${main_gerrit}"
    fi

    if [[ "${LOCAL_HOST}" =~ ^10\.63\.237\. ]] || \
       [[ "${LOCAL_HOST}" =~ ^10\.63\.231\. ]] || \
       [[ "${LOCAL_HOST}" =~ ^10\.156\.196\. ]] || \
       [[ "${LOCAL_HOST}" =~ ^10\.67\.159\. ]]; then
        if [[ "${JARVIS_VERSION_SHEET}" == "true" ]] && \
           [[ "${IS_TEMP_SHEET}" == "false" ]]; then
            local_gerrit="${main_gerrit}"
        fi
    fi

    local strict_gerrit="${STRICT_GERRIT}"
    if [[ -n "${strict_gerrit}" ]]; then
        local_gerrit="${strict_gerrit}"
    fi

    echo ${local_gerrit}
}

# Check environment
ret=$(check_environment)
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    return
fi

# Set location to chengdu
ret=$(go run test.go --sites "10.75.200.210" | tail -n 1 | sed 's/^best site: //')
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    return
fi

# Set location to shanghai
ret=$(go run test.go --sites "10.67.40.202,10.63.237.206,10.67.16.29" | tail -n 1 | sed 's/^best site: //')
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    return
fi

# Set location to xian
ret=$(go run test.go --sites "10.95.243.159,10.95.243.158" | tail -n 1 | sed 's/^best site: //')
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    return
fi
