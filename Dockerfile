FROM ubuntu:22.04
COPY docker-init.sh /usr/local/bin/

# Entrypoint will be set by docker-compose:"command"
ENTRYPOINT []

