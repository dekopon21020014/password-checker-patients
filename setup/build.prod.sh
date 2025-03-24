#!/bin/bash
docker build --platform=linux/amd64 -f ./Dockerfile -t password-checker:amd64 ../src
docker save -o password-checker.tar password-checker:amd64
