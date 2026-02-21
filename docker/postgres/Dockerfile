#--------------------------------------------------------------------------
#
# Imagineer - TTRPG Campaign Intelligence Platform
#
# Copyright (c) 2025 - 2026
# This software is released under The MIT License
#
#--------------------------------------------------------------------------

FROM ghcr.io/pgedge/pgedge-postgres:18-spock5-standard

# Build pgedge_vectorizer from source (beta3)
USER root
COPY --from=vectorizer . /tmp/pgedge-vectorizer
RUN dnf install -y gcc clang llvm-devel make \
        pgedge-postgresql18-devel \
        redhat-rpm-config \
        libcurl-devel && \
    cd /tmp/pgedge-vectorizer && \
    make clean && make && make install && \
    dnf remove -y gcc clang llvm-devel make \
        pgedge-postgresql18-devel \
        redhat-rpm-config && \
    dnf clean all && \
    rm -rf /tmp/pgedge-vectorizer

USER postgres
