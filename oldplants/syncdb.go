package main

import (
	"crypto/sha512"
	"dbsync/models"
	"flag"
	"fmt"
	"log"

	"gorm.io/gorm/clause"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func dropTables(db *gorm.DB) {
	// Drop tables if they exist
	fmt.Println("Dropping tables...")
	if err := db.Migrator().DropTable(
		&models.PlantLogModel{},
		&models.PlantModel{},
		&models.CommentModel{},
		&models.ImageModel{}); err != nil {
		log.Printf("Error dropping tables: %v", err)
	} else {
		log.Println("Dropped tables successfully")
	}
}

func main() {
	srcDbPassword := flag.String("srcDbPassword", "srcPassword", "Source DB Password")
	srcDbUrlName := flag.String("srcDbUrlName", "srcName", "Source DB URL name")
	srcDbName := flag.String("srcDbName", "srcDbName", "Source DB name")
	srcDbUsername := flag.String("srcDbUsername", "srcDbUsername", "Source DB username")
	dstDbPassword := flag.String("dstDbPassword", "dstPassword", "Destination DB Password")
	dstDbUrlName := flag.String("dstDbUrlName", "dstName", "Destination DB URL name")
	dstDbName := flag.String("dstDbName", "dstDbName", "Destination DB name")
	dstDbUsername := flag.String("dstDbUsername", "dstDbUsername", "Destination DB username")
	flag.Parse()

	// parse the password from each database from arguments
	postgresDSN := fmt.Sprintf("host=%s.postgres.database.azure.com user=%s password=%s dbname=%s port=5432 sslmode=require", *srcDbUrlName, *srcDbUsername, *srcDbPassword, *srcDbName)
	sourceDb, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Connect to Azure SQL database
	azureSQLDSN := fmt.Sprintf("server=%s.database.windows.net;user id=%s;password=%s;port=1433;database=%s;", *dstDbUrlName, *dstDbUsername, *dstDbPassword, *dstDbName)
	targetDb, err := gorm.Open(sqlserver.Open(azureSQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Azure SQL: %v", err)
	}

	dropTables(targetDb) // Drop existing tables in the target database to ensure a clean migration

	migratePlants(sourceDb, targetDb)
	migrateComments(sourceDb, targetDb)
	migratePlantLogs(sourceDb, targetDb)
	//migrateImages(sourceDb, targetDb)

}

func migratePlantLogs(sourceDb *gorm.DB, targetDb *gorm.DB) {
	// AutoMigrate the schema for Azure SQL
	if err := targetDb.AutoMigrate(&models.PlantLogModel{}); err != nil {
		log.Fatalf("Failed to migrate Azure SQL schema: %v", err)
	}

	// Fetch all records from PostgreSQL
	var records []models.PlantLogModel
	if err := sourceDb.Find(&records).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	fmt.Println("Found ", len(records), "records in PostgreSQL.")

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

func migrateComments(sourceDb *gorm.DB, targetDb *gorm.DB) {
	// AutoMigrate the schema for Azure SQL
	if err := targetDb.AutoMigrate(&models.CommentModel{}); err != nil {
		log.Fatalf("Failed to migrate Azure SQL schema: %v", err)
	}

	// Fetch all records from PostgreSQL
	var records []models.CommentModel
	if err := sourceDb.Find(&records).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	fmt.Println("Found ", len(records), "records in PostgreSQL.")

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

func migrateImages(sourceDb *gorm.DB, targetDb *gorm.DB) {

	// AutoMigrate the schema for Azure SQL
	if err := targetDb.AutoMigrate(&models.ImageModel{}); err != nil {
		log.Fatalf("Failed to migrate Azure SQL schema: %v", err)
	}

	// Fetch all records from PostgreSQL
	var records []models.ImageModel
	if err := sourceDb.Find(&records).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	fmt.Println("Found ", len(records), "records in PostgreSQL.")

	// store imageId -> sha512 hash of .Data field map
	srcImageIdHashMap := make(map[string]uint)
	for _, record := range records {
		// compute sha512 hash over record.Data
		var hash = sha512.Sum512(record.Data)
		// turn the hash into a string
		hashString := fmt.Sprintf("%x", hash)
		srcImageIdHashMap[hashString] = record.ID
		fmt.Println("Image ID: ", record.ID, " Image Hash: ", hashString)
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

func migratePlants(sourceDb *gorm.DB, targetDb *gorm.DB) {

	// build a plantId -> plantName map from the source db
	srcPlantNameIdMap := make(map[string]uint)
	var plants []models.PlantModel
	if err := sourceDb.Find(&plants).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	for _, plant := range plants {
		fmt.Println("Plant ID: ", plant.ID, " Plant Name: ", plant.Name)
		srcPlantNameIdMap[plant.Name] = plant.ID
	}

	fmt.Println("Connected to PostgreSQL and Azure SQL databases successfully!")

	// AutoMigrate the schema for Azure SQL
	if err := targetDb.AutoMigrate(&models.PlantModel{}); err != nil {
		log.Fatalf("Failed to migrate Azure SQL schema: %v", err)
	}

	// Fetch all records from PostgreSQL
	var records []models.PlantModel
	if err := sourceDb.Find(&records).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	fmt.Println("Found ", len(records), "records in PostgreSQL.")

	// Synchronize data to Azure SQL
	for _, record := range records {
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Fatalf("Failed to upsert record ID %d: %v", record.ID, err)
		}
		// fmt.Println("Will change plant ID: ", record.ID, " Plant Name: ", record.Name, " to ID: ", srcPlantNameIdMap[record.Name])
		// record.ID = srcPlantNameIdMap[record.Name]
		// targetDb.Save(&record)
	}
	fmt.Printf("Synchronization complete, migrated %d plants!\n", len(records))

	var plantsDst []models.PlantModel
	if err := targetDb.Find(&plantsDst).Error; err != nil {
		log.Fatalf("Failed to fetch records from target db: %v", err)
	}
	for _, plant := range plantsDst {
		fmt.Println("Would change plant ID: ", plant.ID, " Plant Name: ", plant.Name, " to ID: ", srcPlantNameIdMap[plant.Name])
		// plant.ID = srcPlantNameIdMap[plant.Name]
		// targetDb.Save(&plant)
	}
}
