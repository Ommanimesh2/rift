#!/bin/bash
# Produce demo.gif for README
set -e
echo "Pulling demo images into Docker daemon..."
docker pull nginx:1.24 nginx:1.25 ubuntu:22.04 ubuntu:24.04
echo "Recording demo with vhs..."
vhs demo.tape
echo "Done: demo.gif ($(du -h demo.gif | cut -f1))"
