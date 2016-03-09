#!/usr/bin/env bash

set -e

path=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cli="$path/../../../bin/esxcloud-cli"

# Store initial list of tenants
initialTenants=$($cli -n tenant list | awk '{print $2}')

# Create many tenants named by a letter in the alphabet
for name in {a..z}
do
    $cli -n tenant create $name
done

# Retrieve an id and check for CREATE_TENANT operationls
id=$($cli -n tenant list | awk '{if (NR==3) {print $2}}')
operation=$($cli -n tenant tasks $id | awk '{if (NR==3) {print $2}}')
if [ "CREATE_TENANT" != $operation ]
then
    exit 1
fi

# List out all the tenants and delete tenants not in the initial list
output=$($cli -n tenant list | awk '{if (NR!=1) {print $2}}')
for line in $output
do
    if ! [[ ${initialTenants[*]} =~ $line ]]
    then
        $cli -n tenant delete $line
    fi
done
