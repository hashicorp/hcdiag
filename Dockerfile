# Copyright IBM Corp. 2021, 2025
# SPDX-License-Identifier: MPL-2.0

FROM ubuntu:22.04
COPY docker-init.sh /usr/local/bin/

# Entrypoint will be set by docker-compose:"command"
ENTRYPOINT []

