#!/usr/bin/env bash

set -e

# Get path of executable binary file
path=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cli="$path/../../../bin/esxcloud-cli"

# Set target with hardcoded endpoint
$cli -n target set "http://localhost:9080"

# Save initial list of images
initialImages=$($cli -n image list | awk '{print $1}')

# Create a new image
echo "Creating test image"
id=$($cli -n image upload "../../testdata/tty_tiny.ova" -n "testname" -i "EAGER" | awk '{if (NR != 1) {print $1}}')

# Retrieve image create state
state=$($cli -n image show $id | awk '{if (NR != 1) {print $3}}')

# Verify image state is ready
if [ $state != "READY" ]
then
    echo "Error: image created not ready"
    exit 1
fi

# Delete image by id
echo "Deleting test image"
$cli -n image delete $id > /dev/null

# Verify images was deleted
output=$($cli -n tenant list | awk '{if (NR!=1) {print $2}}')
for line in $output
do
    if [[ ${initialImages[*]} =~ $line ]]
    then
        echo "Error: Image $id should be deleted"
        exit 1
    fi
done
