RG_NAME="akswebapprg"
ACRNAME="mcoronistestacr3"

# acr steps
# az acr build --registry $ACRNAME --image aks-test/test_frontend:latest ../../docker/frontend/
az acr build --registry $ACRNAME --image aks-test/plantmindr-frontend:latest ../../docker/plantmindr/
az acr build --registry $ACRNAME --image aks-test/plantmindr-backend:latest ../../docker/backend/
az acr repository list --name $ACRNAME --output table