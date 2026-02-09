#--------------------------------------------------------------------------
#
# Imagineer - TTRPG Campaign Intelligence Platform
#
# Copyright (c) 2025 - 2026
# This software is released under The MIT License
#
#--------------------------------------------------------------------------

FROM ollama/ollama:latest
RUN ollama serve & sleep 5 && ollama pull mxbai-embed-large \
    && kill %1
