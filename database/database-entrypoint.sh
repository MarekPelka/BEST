#!/bin/bash

# Backup:
# #su root -s /bin/bash -m -c 'set -o monitor && /opt/mssql/bin/sqlservr & sleep 10 && /opt/mssql-tools/bin/sqlcmd -S 127.0.0.1\\SQLEXPRESS,1433 -U sa -P Microsoft2017 -Q "CREATE DATABASE SampleDB;" && /opt/mssql-tools/bin/sqlcmd -S 127.0.0.1\\SQLEXPRESS,1433 -U sa -P Microsoft2017 -Q "SELECT name FROM master.dbo.sysdatabases" && jobs && fg %1'
# pkill sqlservr 

# Start the database process
/opt/mssql/bin/sqlservr &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start my_first_process: $status"
  exit $status
fi

sleep 12
/opt/mssql-tools/bin/sqlcmd -S 127.0.0.1\\SQLEXPRESS,1433 -U sa -P Microsoft2017 -Q "CREATE DATABASE SampleDB;"
/opt/mssql-tools/bin/sqlcmd -S 127.0.0.1\\SQLEXPRESS,1433 -U sa -P Microsoft2017 -d SampleDB -i /workspace/initial.sql
/opt/mssql-tools/bin/sqlcmd -S 127.0.0.1\\SQLEXPRESS,1433 -U sa -P Microsoft2017 -Q "SELECT name FROM master.dbo.sysdatabases"

while sleep 60; do
  ps aux | grep sqlservr | grep -q -v grep
  PROCESS_1_STATUS=$?

  if [ $PROCESS_1_STATUS -ne 0 ]; then
    echo "One of the processes has already exited."
    exit 1
  else
    echo "[$(date)] Still running."
    #pkill sqlservr 
  fi
done

