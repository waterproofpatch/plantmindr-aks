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

	dropTables(targetDb)

	var _ = migratePlants(sourceDb, targetDb)
	// migrateComments(sourceDb, targetDb, plantIdMap)
	// migratePlantLogs(sourceDb, targetDb, plantIdMap)
	log.Printf("Migrating plant images...")
	var imageIdMap = migrateImages(sourceDb, targetDb)

	// now update the plant records with the new imageId
	var plants []models.PlantModel
	if err := targetDb.Find(&plants).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	fmt.Println("Found ", len(plants), "records in PostgreSQL.")
	for _, plant := range plants {
		var newImageId = imageIdMap[plant.ImageId]
		fmt.Println("Updating plant old imageId: ", plant.ImageId, " Image ID: ", newImageId)
		plant.ImageId = newImageId
		if err := targetDb.Save(&plant).Error; err != nil {
			log.Printf("Failed to upsert record ID %d: %v", plant.ID, err)
		}
	}

}

func migratePlantLogs(sourceDb *gorm.DB, targetDb *gorm.DB, plantIdMap map[uint]uint) {
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
		record.PlantID = int(plantIdMap[uint(record.PlantID)])
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Printf("Failed to upsert record ID %d: %v", record.ID, err)
		}
	}
	fmt.Println("Synchronization complete!")
}

func migrateComments(sourceDb *gorm.DB, targetDb *gorm.DB, plantIdMap map[uint]uint) {
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
		record.PlantID = int(plantIdMap[uint(record.PlantID)])
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Printf("Failed to upsert record ID %d: %v", record.ID, err)
		}
	}
	fmt.Println("Synchronization complete!")
}

func migrateImages(sourceDb *gorm.DB, targetDb *gorm.DB) map[int]int {

	// AutoMigrate the schema for Azure SQL
	if err := targetDb.AutoMigrate(&models.ImageModel{}); err != nil {
		log.Fatalf("Failed to migrate Azure SQL schema: %v", err)
	}

	var idMap = make(map[int]int)

	// Fetch all records from PostgreSQL
	var records []models.ImageModel
	if err := sourceDb.Find(&records).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	fmt.Println("Found ", len(records), "records in PostgreSQL.")

	// Synchronize data to Azure SQL
	for _, record := range records {
		fmt.Println("Old image.ID: ", record.ID)
		var oldId = int(record.ID)
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Printf("Failed to upsert record ID %d: %v", record.ID, err)
		}
		fmt.Println("New image.ID: ", record.ID)
		idMap[oldId] = int(record.ID)
	}
	fmt.Println("Synchronization complete!")
	return idMap
}

// returns a map of previous plantId to new plantId
func migratePlants(sourceDb *gorm.DB, targetDb *gorm.DB) map[uint]uint {

	fmt.Println("Connected to PostgreSQL and Azure SQL databases successfully!")
	var plantIdMap = make(map[uint]uint)

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
		var oldId = record.ID
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Fatalf("Failed to upsert record ID %d: %v", record.ID, err)
		}
		var newId = record.ID
		plantIdMap[oldId] = newId
	}
	fmt.Printf("Synchronization complete, migrated %d plants!\n", len(records))
	return plantIdMap
}
