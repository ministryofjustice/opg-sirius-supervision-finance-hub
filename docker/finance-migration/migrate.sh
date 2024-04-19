#! /usr/bin/env sh

set -e
set -o pipefail

DB_CONN_STR="postgres://$DB_USER:$DB_PASSWORD@$DB_CONNECTION/$DB_NAME?search_path=$DB_SCHEMA"
COMMAND=$1

check_var() {
    NAME=$1
    VAR=$2
    if [ -z "$VAR" ]; then
        echo "ERROR - Make sure $NAME is set."
        exit 1
    fi
}

check_var DB_USER $DB_USER
check_var DB_PASSWORD $DB_PASSWORD
check_var DB_CONNECTION $DB_CONNECTION
check_var DB_NAME $DB_NAME
check_var DB_SCHEMA $DB_SCHEMA
check_var COMMAND $1

/usr/local/bin/goose -dir /database postgres $DB_CONN_STR $COMMAND
