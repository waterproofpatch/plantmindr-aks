RG_NAME="akswebapprg"
CLUSTER_NAME="webappakscluster"

# kubectl steps
echo "Getting credentials for the cluster..."
az aks get-credentials --resource-group $RG_NAME --name $CLUSTER_NAME
kubectl get nodes