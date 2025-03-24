RG_NAME="akswebapprg"
CLUSTER_NAME="webappakscluster"
LOCATION="eastus"
KEY_NAME="sshKey"
DNS_PREFIX="mcoronistestaks"
USER="azureuser"
ACRNAME="mcoronistestacr3"

RG_DEPLOY_BICEP="rg.bicep"
AKS_DEPLOY_BICEP="aks.bicep"

az login

echo "Creating RG deployment..."
az deployment sub create \
    --name testdeployment \
    --location $LOCATION \
    --template-file $RG_DEPLOY_BICEP \
    --parameters resourceGroupName=$RG_NAME resourceGroupLocation=$LOCATION

# Create an SSH key pair using Azure CLI. Fails noncrit if key exists
az sshkey create --name "$KEY_NAME" --resource-group $RG_NAME
PUBLIC_KEY=$(az sshkey show --name "$KEY_NAME" --resource-group $RG_NAME --query "publicKey" --output tsv)

# # aks cluster (rg level)
echo "Creating AKS deployment..."
az deployment group create \
    --resource-group $RG_NAME \
    --template-file $AKS_DEPLOY_BICEP \
    --parameters \
    clusterName=$CLUSTER_NAME \
    dnsPrefix=$DNS_PREFIX \
    linuxAdminUsername=$USER \
    acrName=$ACRNAME \
    sshRSAPublicKey="$PUBLIC_KEY"

echo "Getting AKS credentials..."
az aks get-credentials --resource-group $RG_NAME --name $CLUSTER_NAME