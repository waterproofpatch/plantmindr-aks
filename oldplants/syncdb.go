package main

import (
	"dbsync/models"
	"flag"
	"fmt"
	"log"

	"gorm.io/gorm/clause"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func main() {
	srcDbPassword := flag.String("srcDbPassword", "srcPassword", "Source DB Password")
	srcDbUrlName := flag.String("srcDbUrlName", "srcName", "Source DB URL name")
	srcDbName := flag.String("srcDbName", "srcDbName", "Source DB name")
	dstDbPassword := flag.String("dstDbPassword", "dstPassword", "Destination DB Password")
	dstDbUrlName := flag.String("dstDbUrlName", "dstName", "Destination DB URL name")
	dstDbName := flag.String("dstDbName", "dstDbName", "Destination DB name")
	flag.Parse()

	// parse the password from each database from arguments
	postgresDSN := fmt.Sprintf("host=%s.postgres.database.azure.com user=freeman password=%s dbname=%s port=5432 sslmode=require", *srcDbUrlName, *srcDbPassword, *srcDbName)
	sourceDb, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Connect to Azure SQL database
	azureSQLDSN := fmt.Sprintf("server=%s.database.windows.net;user id=plantmindr;password=%s;port=1433;database=%s;", *dstDbUrlName, *dstDbPassword, *dstDbName)
	targetDb, err := gorm.Open(sqlserver.Open(azureSQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Azure SQL: %v", err)
	}

	fmt.Println("Connected to PostgreSQL and Azure SQL databases successfully!")

	fmt.Println("Migrating PlantModel...")
	// AutoMigrate the schema for Azure SQL
	if err := targetDb.AutoMigrate(&models.PlantModel{}); err != nil {
		log.Fatalf("Failed to migrate Azure SQL schema: %v", err)
	}

	// Fetch all records from PostgreSQL
	var records []models.PlantModel
	if err := sourceDb.Find(&records).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}

	// Synchronize data to Azure SQL
	for _, record := range records {
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Printf("Failed to upsert record ID %d: %v", record.ID, err)
		}
	}
	fmt.Println("Synchronization complete!")
}
