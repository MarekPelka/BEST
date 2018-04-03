#!/bin/bash
set -x #debug
HOST=127.0.0.1
PORT=1401
USER=sa
PASSWORD=Microsoft2017
DATABASE=SampleDB

displayUsage() {
    echo "Usage:"
    echo "$0                    -- T-sql command line"
    echo "$0 "query"              -- Query to be run on DB"
    echo "$0 datebase file.sql  -- File to be run on DB" 
    exit 0
}

if [ $# -eq 0 ]; then
    OPTS=""
elif [ $# -eq 1 ]; then
    if [[ $1 = *".sql"* ]] && [ -n $DATABASE ]; then
        OPTS="-d $DATABASE -i ./$1"
    else
        BASEOPTS="-Q $1"
    fi
elif [ $# -eq 2 ]; then
    if [[ $2 = *".sql"* ]]; then
        OPTS="-d $1 -i ./$2"
    else 
        displayUsage
    fi
else
    displayUsage
fi
sqlcmd -S $HOST\\SQLEXPRESS,$PORT -U $USER -P $PASSWORD $OPTS

