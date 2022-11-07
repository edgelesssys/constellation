#!/usr/bin/env bash

helm pull cert-manager --version 1.10.0 --repo https://charts.jetstack.io --untar --untardir charts && rm -rf charts/cert-manager/README.md charts/cert-manager-v1.10.0.tgz
git apply ./cert-manager.patch
