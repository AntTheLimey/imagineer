#--------------------------------------------------------------------------
#
# Imagineer - TTRPG Campaign Intelligence Platform
#
# Copyright (c) 2025 - 2026
# This software is released under The MIT License
#
#--------------------------------------------------------------------------

FROM ollama/ollama:latest
RUN ollama serve & OLLAMA_PID=$! && sleep 5 && ollama pull mxbai-embed-large; \
    kill $OLLAMA_PID 2>/dev/null; exit 0
