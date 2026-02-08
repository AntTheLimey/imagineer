#!/usr/bin/env bash
#--------------------------------------------------------------------------
#
# Imagineer - TTRPG Campaign Intelligence Platform
#
# Copyright (c) 2025 - 2026
# This software is released under The MIT License
#
#--------------------------------------------------------------------------
set -euo pipefail

# Pull the embedding model used by pgedge_vectorizer.
# This is idempotent -- Ollama skips the download if the
# model is already present in the ollama-models volume.

echo "Pulling mxbai-embed-large embedding model..."
docker exec imagineer-ollama ollama pull mxbai-embed-large
echo "Model ready."
