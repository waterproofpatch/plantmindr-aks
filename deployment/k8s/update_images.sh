#!/bin/bash

# Parameters
RG_NAME="akswebapprg"
ACRNAME="mcoronistestacr3"
IMAGE_NAME=$1

if [ -z "$IMAGE_NAME" ]; then
  echo "Usage: $0 <image_name>"
  exit 1
fi

# ACR steps
case $IMAGE_NAME in
  frontend)
    az acr build --registry $ACRNAME --image aks-test/plantmindr-frontend:latest ../../docker/plantmindr/
    ;;
  backend)
    az acr build --registry $ACRNAME --image aks-test/plantmindr-backend:latest ../../docker/backend/
    ;;
  *)
    echo "Unknown image name: $IMAGE_NAME. Valid options are 'frontend' or 'backend'."
    exit 1
    ;;
esac

az acr repository list --name $ACRNAME --output table
