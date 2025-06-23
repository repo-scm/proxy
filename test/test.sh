#!/bin/bash

LOCAL_HOST=""

MAIN_GERRIT="10.67.16.29"
STRICT_GERRIT="10.67.16.29"

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

    return local_gerrit
}

# Check environment
check_environment
ret=$?
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    exit 0
fi

# Set location to chengdu
ret=$(./bin/test --sites "10.75.200.210")
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    exit 0
fi

# Set location to shanghai
ret=$(./bin/test --sites "10.67.40.202,10.63.237.206,10.67.16.29")
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    exit 0
fi

# Set location to xian
ret=$(./bin/test --sites "10.95.243.159,10.95.243.158")
if [[ -n "${ret}" ]]; then
    echo "LOCAL_GERRIT=${ret}"
    export LOCAL_GERRIT="${ret}"
    exit 0
fi

exit 0
