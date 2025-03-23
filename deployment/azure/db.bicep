// param serverName string
// param adminUser string
// param adminPassword string
// param location string = resourceGroup().location

// resource postgresServer 'Microsoft.DBforPostgreSQL/flexibleServers@2021-06-01' = {
//   name: serverName
//   location: location
//   sku: {
//     name: 'Standard_D2s_v3'
//     tier: 'GeneralPurpose'
//   }
//   properties: {
//     administratorLogin: adminUser
//     administratorLoginPassword: adminPassword
//     version: '13'
//     storage: {
//       storageSizeGB: 32
//     }
//     highAvailability: {
//       mode: 'Disabled'
//     }
//     backup: {
//       backupRetentionDays: 7
//       geoRedundantBackup: 'Disabled'
//     }
//   }
// }

// output serverFQDN string = postgresServer.properties.fullyQualifiedDomainName
