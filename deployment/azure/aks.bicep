param location string = resourceGroup().location
param acrName string
param clusterName string
param dnsPrefix string
param osDiskSizeGB int = 0
param agentCount int = 1
param agentVMSize string = 'standard_d2s_v3'
param linuxAdminUsername string
param sshRSAPublicKey string

resource aks 'Microsoft.ContainerService/managedClusters@2024-02-01' = {
  name: clusterName
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    dnsPrefix: dnsPrefix
    agentPoolProfiles: [
      {
        // name: 'agentpool'
        name: 'spotpool'
        osDiskSizeGB: osDiskSizeGB
        count: agentCount
        vmSize: agentVMSize
        osType: 'Linux'
        mode: 'System'
        spotMaxPrice: -1
      }
    ]
    linuxProfile: {
      adminUsername: linuxAdminUsername
      ssh: {
        publicKeys: [
          {
            keyData: sshRSAPublicKey
          }
        ]
      }
    }
  }
}

resource acr 'Microsoft.ContainerRegistry/registries@2023-11-01-preview' = {
  name: acrName
  location: location
  sku: {
    name: 'Standard'
  }
}

var acrPullRoleDefinitionId = subscriptionResourceId(
  'Microsoft.Authorization/roleDefinitions',
  '7f951dda-4ed3-4680-a7ca-43fe172d538d'
)

resource acrPullRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, aks.id, acrPullRoleDefinitionId)
  scope: acr
  properties: {
    principalId: aks.properties.identityProfile.kubeletidentity.objectId
    roleDefinitionId: acrPullRoleDefinitionId
    principalType: 'ServicePrincipal'
  }
}

resource vault 'Microsoft.KeyVault/vaults@2023-07-01' = {
  name: 'plantmindrrbackv'
  location: location
  properties: {
    sku: {
      family: 'A'
      name: 'premium'
    }
    tenantId: tenant().tenantId
    softDeleteRetentionInDays: 90
    enableSoftDelete: true
    enablePurgeProtection: true
    enableRbacAuthorization: true
    networkAcls: {
      defaultAction: 'Allow' // todo: enable public access
      bypass: 'AzureServices'
    }
    publicNetworkAccess: 'Enabled'
  }
}

// var kvSecretsUserRoleDefinitionId = subscriptionResourceId(
//   'Microsoft.Authorization/roleDefinitions',
//   '4633458b-17de-408a-b874-0445c86b69e6' // Key Vault Secrets User role
// )

// resource kvSecretsUserRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
//   name: guid(resourceGroup().id, aks.id, kvSecretsUserRoleDefinitionId)
//   scope: vault
//   properties: {
//     principalId: aks.identity.principalId
//     roleDefinitionId: kvSecretsUserRoleDefinitionId
//     principalType: 'ServicePrincipal'
//   }
// }

output controlPlaneFQDN string = aks.properties.fqdn
output controlPlanePrincipalId string = aks.identity.principalId
