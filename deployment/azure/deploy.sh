RG_NAME="akswebapprg"
CLUSTER_NAME="webappakscluster"
LOCATION="eastus"
KEY_NAME="sshKey"
DNS_PREFIX="mcoronistestaks"
USER="azureuser"
ACRNAME="mcoronistestacr3"
SUB_ID="f9eebc87-0701-4ec6-a34d-25ace368eafd"
KV_MSI_NAME="akskvidentity"
K8_SERVICE_ACCOUNT_NAME="akskvidentityserviceaccount"

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

echo "Enabling OIDC and Workload Identity..."
az aks update -g $RG_NAME -n $CLUSTER_NAME --enable-oidc-issuer --enable-workload-identity
az identity create --name $KV_MSI_NAME --resource-group $RG_NAME  --location $LOCATION --subscription $SUB_ID
USER_ASSIGNED_CLIENT_ID=$(az identity show --resource-group $RG_NAME --name $KV_MSI_NAME --query 'clientId' -otsv)
echo "Client ID: $USER_ASSIGNED_CLIENT_ID"


echo "Getting AKS credentials..."
az aks get-credentials --resource-group $RG_NAME --name $CLUSTER_NAME

AKS_OIDC_ISSUER=$(az aks show -n $CLUSTER_NAME -g $RG_NAME --query "oidcIssuerProfile.issuerUrl" -otsv)
echo "AKS OIDC Issuer: $AKS_OIDC_ISSUER"

echo """
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    azure.workload.identity/client-id: "$USER_ASSIGNED_CLIENT_ID"
  name: "$K8_SERVICE_ACCOUNT_NAME"
  namespace: "default"
  labels:
    azure.workload.identity/use: \"true\"
""" > identity-service-account.yaml

kubectl apply -f identity-service-account.yaml

IDENTITY_TENANT=$(az aks show --name $CLUSTER_NAME --resource-group $RG_NAME --query identity.tenantId -o tsv)

# TODO: unsure if needed
cat <<EOF | kubectl apply -f -
# This is a SecretProviderClass example using workload identity to access your key vault
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: azure-kvname-wi # needs to be unique per namespace
spec:
  provider: azure
  parameters:
    usePodIdentity: "false"
    clientID: "$USER_ASSIGNED_CLIENT_ID" # Setting this to use workload identity
    keyvaultName: plantmindrrbackv       # Set to the name of your key vault
    cloudName: ""                         # [OPTIONAL for Azure] if not provided, the Azure environment defaults to AzurePublicCloud
    objects:  |
      array:
        - |
          objectName: secret1             # Set to the name of your secret
          objectType: secret              # object types: secret, key, or cert
          objectVersion: ""               # [OPTIONAL] object versions, default to latest if empty
        - |
          objectName: key1                # Set to the name of your key
          objectType: key
          objectVersion: ""
    tenantId: "${IDENTITY_TENANT}"        # The tenant ID of the key vault
EOF

az identity federated-credential create --name "akskvfederatedidentity" --identity-name $KV_MSI_NAME --resource-group $RG_NAME --issuer $AKS_OIDC_ISSUER --subject system:serviceaccount:"default:$K8_SERVICE_ACCOUNT_NAME" --audience api://AzureADTokenExchange
