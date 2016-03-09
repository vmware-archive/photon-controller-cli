#!/usr/bin/env bash

set -e

# Get path of executable binary file
path=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cli="$path/../../../bin/esxcloud-cli"

# Function to parse json string in the format of k1:v1,k2:v2...
# take two input: string variable and key
parse_json () {
    echo $1 | awk -F'[,:]+' -v key="$2" '{for (i=1;i<NF;i++) if ($i==key) print $(i+1)}'
}

# Set target with hardcoded endpoint
$cli -n target set "http://localhost:9080"

# Store initial list of host id
# host list scripting output:
#   Using target ...
#   num_hosts
#   host1_id host1_ip host1_tag host1_state
#   host2_id host2_ip host2_tag host2_state
#   ...
output=$($cli -n host list)
initialHosts=$(echo "$output" | sed 1,2d | awk '{print $1}')

# Test creating a MGMT host
# host create scripting output:
#   Using target ...
#   host_id
echo "Creating 1 MGMT host"
output=$($cli -n host create -u "u" -p "p" -t "MGMT" -i "119.170.211.85" \
            -m "{\"MANAGEMENT_DATASTORE\":\"test_datastore\",
                \"MANAGEMENT_NETWORK_DNS_SERVER\":\"test_server\",
                \"MANAGEMENT_NETWORK_GATEWAY\":\"test_gateway\",
                \"MANAGEMENT_NETWORK_IP\":\"test_ip\",
                \"MANAGEMENT_NETWORK_NETMASK\":\"test_netmask\",
                \"MANAGEMENT_PORTGROUP\":\"test_portgroup\"}")
mgmtHostID=$(echo "$output" | sed 1d)

# Test showing the MGMT host info
# host show scripting output:
#   Using target ...
#   host_id host_ip host_tag host_state host_metadata
echo "Checking the MGMT host info"
output=$($cli -n host show $mgmtHostID)
mgmtHostState=$(echo "$output" | sed 1d | awk '{print $4}')
mgmtHostMetadata=$(echo "$output" | sed 1d | awk '{print $5}')

# Test if the host is created successfully
if [ "$mgmtHostState" != "READY" ]; then
    echo "Error: host create failed"
    exit 1
fi

# Test if metadata in json string format is parseable
# metadata scripting output:
#   k1:v1,k2:v2,k3:v3...
MANAGEMENT_NETWORK_IP=$(parse_json $mgmtHostMetadata "MANAGEMENT_NETWORK_IP")
if [ "$MANAGEMENT_NETWORK_IP" != "test_ip" ]; then
    echo "Error: metadata in 'host show' different from specified"
    exit 1
fi

# Test deleting the MGMT host
# host delete scripting output:
#   Using target ...
#   host_id
echo "Deleting the MGMT host"
$cli -n host delete $mgmtHostID > /dev/null

# Test creating many CLOUD hosts
echo "Creating 10 CLOUD hosts"
for i in {1..10}; do
     $cli -n host create -u "u" -p "p" -t "CLOUD" -i "$i.$i.$i.$i" > /dev/null
done

# Test listing all hosts and deleting hosts not in the initial list
echo "Deleting 10 CLOUD hosts"
output=$($cli -n host list)
idlist=$(echo "$output" | sed 1,2d | awk '{print $1}')
for id in $idlist; do
    if ! [[ ${initialHosts[*]} =~ $id ]]; then
        $cli -n host delete $id > /dev/null
    fi
done
