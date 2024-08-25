#!/bin/bash
echo "Applying deploment..."
kubectl apply -f deployment.yaml

echo "Creating nginx ingress controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.3.0/deploy/static/provider/cloud/deploy.yaml

echo "Checking for public IP"

# kubectl get service ingress-nginx-controller --namespace=ingress-nginx
EXTERNAL_IP=$(kubectl get service ingress-nginx-controller --namespace=ingress-nginx -o json | jq -r '.status.loadBalancer.ingress[0].ip')
echo "External IP: $EXTERNAL_IP"

echo "Testing site..."
wget http://$EXTERNAL_IP

# get info about ingress
# kubectl get ingress -A
