#!/usr/bin/env bash

TMP_FILE="/tmp/out"

# test whether output of the command contains expected lines
# arguments
# 1-st command to run
# 2-nd array of expected strings in the command output
# 3-rd argument is an optional command runtime limit
function testOutput {
IFS="${PREV_IFS}"

    #run the command
    if [ $# -ge 3 ]; then
        $1 > ${TMP_FILE} 2>&1 &
        CMD_PID=$!
        sleep $3
        kill $CMD_PID
    else
        $1 > ${TMP_FILE} 2>&1
    fi

sleep 2
IFS="
"
    echo "Output of $1:"
    cat ${TMP_FILE}
    echo ""
    echo "Checking output of $1"
    rv=0
    # loop through expected lines
    for i in $2
    do
        if grep "${i}" /tmp/out > /dev/null ; then
            echo "OK - '$i'"
        else
            echo "Not found - '$i'"
            rv=1
        fi
    done

    # if an error occurred print the output
    if [[ ! $rv -eq 0 ]] ; then
        cat ${TMP_FILE}
        exitCode=1
    fi

    echo "================================================================"
    rm ${TMP_FILE}
    return ${rv}
}