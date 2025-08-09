#!/bin/bash
echo "Applying deploment..."
# kubectl apply -f deployment-frontend.yaml
kubectl apply -f deployment-frontend.yaml --validate=false
kubectl apply -f deployment-backend.yaml --validate=false

echo "Creating nginx ingress controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.3.0/deploy/static/provider/cloud/deploy.yaml

echo "Checking for public IP"

EXTERNAL_IP=$(kubectl get service ingress-nginx-controller --namespace=ingress-nginx -o json | jq -r '.status.loadBalancer.ingress[0].ip')
echo "External IP: $EXTERNAL_IP"
