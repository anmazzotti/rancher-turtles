#!/usr/bin/env bash

# Copyright © 2026 SUSE LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script dumps some etcd info.
# Mainly this is used to detect and debug conflicts managing resources.
# A dump containing the list of keys and their versions number is included in the dump.
 
set -e

OUTPUT_DIR=${OUTPUT_DIR:-/tmp/capi-test/etcd}
CLUSTER_NAME=${CLUSTER_NAME:-capi-test}

# Install etcdctl
apt update
apt install etcd-client -y

# Dump human readable status
echo "Dumping endpoint status"
ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/peer.crt --key=/etc/kubernetes/pki/etcd/peer.key endpoint status --write-out=table > $OUTPUT_DIR/$CLUSTER_NAME-status.txt

# Dump keys collection, sorted by versions
echo "Collecting keys... this may take a while"

rm -f /tmp/$CLUSTER_NAME-keys.txt # Clear the intermediate file, if any.

# For each key, dump some info
for key in `etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/peer.crt --key=/etc/kubernetes/pki/etcd/peer.key get --prefix --keys-only /`
do
  size=`etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/peer.crt --key=/etc/kubernetes/pki/etcd/peer.key get $key --print-value-only | wc -c`
  count=`etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/peer.crt --key=/etc/kubernetes/pki/etcd/peer.key get $key --write-out=fields | grep \"Count\" | cut -f2 -d':'`
  if [ $count -ne 0 ]; then
    versions=`etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/peer.crt --key=/etc/kubernetes/pki/etcd/peer.key get $key --write-out=fields | grep \"Version\" | cut -f2 -d':'`
  else
    versions=0
  fi
  total=$(($size * $versions))
  printf "$total\t$size\t$versions\t$count\t$key\n" >> /tmp/$CLUSTER_NAME-keys.txt
done

# Sort keys by versions count
printf "TOTAL\tSIZE\tVERSIONS\tCOUNT\tKEY\n" > $OUTPUT_DIR/$CLUSTER_NAME-keys-sorted.txt
sort -nrk 3 /tmp/$CLUSTER_NAME-keys.txt >> $OUTPUT_DIR/$CLUSTER_NAME-keys-sorted.txt

# Exit with error if size exceeded
echo "Taking snapshot to calculate database size"
ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/peer.crt --key=/etc/kubernetes/pki/etcd/peer.key snapshot save /tmp/$CLUSTER_NAME-etcd-snapshot.db
size=$(stat -c%s /tmp/$CLUSTER_NAME-etcd-snapshot.db)
size_limit=500000000

rm -f /tmp/$CLUSTER_NAME-etcd-snapshot.db # Clear some space. No longer needed.

if (( size > size_limit )); then
  printf "ETCD database size exceeding limit of 500MB. Found: $size bytes\n" >&2
  exit 1
fi 
