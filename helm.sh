#!/bin/bash
helm template \
  galvanic chart/openfaas/ \
  --namespace galvanic \
  --set basic_auth=false \
  --set prometheus.create=false \
  --set alertmanager.create=false \
  --set faasIdler.create=false \
  --set functionNamespace=galvanic-fn > galvanic.yaml
