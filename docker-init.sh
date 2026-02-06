#!/usr/bin/env bash
# Copyright IBM Corp. 2021, 2025
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# Install dependencies
apt-get update && apt-get install -y wget gpg

# Add Hashicorp repo
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor > /usr/share/keyrings/hashicorp-archive-keyring.gpg && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com jammy main" | tee /etc/apt/sources.list.d/hashicorp.list

# Install binaries
apt-get update && apt-get install -y nomad consul

# Vault post-install script is not smart enough to work inside of containers that don't give access to setcap
# e.g.:
# /var/lib/dpkg/info/vault.postinst: line 36: setcap: command not found
set +e
apt-get install -y vault
set -e

# Move
mkdir -p /hashistack && cd /hashistack

# Start services
#
# no ACLs for now
# echo "acl { enabled = true }" > nomad.hcl
#
# nomad agent -dev -config=nomad.hcl  >> nomad.log 2>&1 &
nomad agent -dev >> nomad.log 2>&1 &
consul agent -dev >> consul.log 2>&1 &
vault server -dev  >> vault.log 2>&1 &

# Shell configuration niceties (so you can run the products without errors on the CLI)
echo "export CONSUL_HTTP_ADDR=http://127.0.0.1:8500" >> hashi.env
echo "export VAULT_ADDR=http://127.0.0.1:8200" >> hashi.env

# Instructions
echo "Setup complete. Get started with the following commands:"
echo
echo "1. Compile hcdiag:"
echo "GOOS=linux GOARCH=amd64 go build -o bin/hcdiag"
echo
echo "2. Connect to the container in a new shell:"
echo "'docker exec -it hcdiag-hashistack-1 /bin/bash'"
echo
echo "3. Move to the working directory and source the env file:"
echo "cd /hashistack && source hashi.env"
echo

# Run forever
echo "About to run forever..."
tail -F runforever

