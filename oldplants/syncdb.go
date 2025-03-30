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
	srcDbName := flag.String("srcDbName", "srcName", "Source DB name (postgres)")
	dstDbPassword := flag.String("dstDbPassword", "dstPassword", "Destination DB Password")
	dstDbName := flag.String("dstDbName", "dstName", "Destination DB name (postgres)")
	flag.Parse()
	fmt.Printf("Source DB Password: %s\n", *srcDbPassword)
	fmt.Printf("Destination DB Password: %s\n", *dstDbPassword)
	// return

	// parse the password from each database from arguments
	postgresDSN := fmt.Sprintf("host=%s.postgres.database.azure.com user=freeman password=%s dbname=postgres port=5432 sslmode=require", *srcDbName, *srcDbPassword)
	sourceDb, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Connect to Azure SQL database
	azureSQLDSN := fmt.Sprintf("server=%s.database.windows.net;user id=plantmindr;password=%s;port=1433;database=plantmindersqldb;", *dstDbName, *dstDbPassword)
	targetDb, err := gorm.Open(sqlserver.Open(azureSQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Azure SQL: %v", err)
	}

	fmt.Println("Connected to PostgreSQL and Azure SQL databases successfully!")

	fmt.Println("Starting synchronization...")

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
