az deployment group create --resource-group "akswebapprg" --template-file db.bicep --parameters serverName="plantmindrdbserver" adminUser="adminUser" adminPassword="adminPassword"

