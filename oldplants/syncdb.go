package main

import (
	"crypto/sha256"
	"dbsync/models"
	"flag"
	"fmt"
	"log"
	"os"

	"gorm.io/gorm/clause"

	"gorm.io/driver/postgres"
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

func getImageRecords(db *gorm.DB) []models.ImageModel {
	var images []models.ImageModel
	if err := db.Find(&images).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	return images
}

func getPlantRecords(db *gorm.DB) []models.PlantModel {
	var plants []models.PlantModel
	if err := db.Find(&plants).Error; err != nil {
		log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	}
	return plants
}

func main() {
	// dropOnly := flag.Bool("drop", false, "If set, only drop tables and exit")
	flag.Parse()

	// srcDbPassword := os.Getenv("SRC_DB_PASSWORD")
	// srcDbUrlName := os.Getenv("SRC_DB_URL_NAME")
	// srcDbName := os.Getenv("SRC_DB_NAME")
	// srcDbUsername := os.Getenv("SRC_DB_USERNAME")
	// dstDbPassword := os.Getenv("DST_DB_PASSWORD")
	// dstDbUrlName := os.Getenv("DST_DB_URL_NAME")
	// dstDbName := os.Getenv("DST_DB_NAME")
	// dstDbUsername := os.Getenv("DST_DB_USERNAME")
	srcDbConnString := os.Getenv("SRC_DB_CONN_STRING")
	dstDbConnString := os.Getenv("DST_DB_CONN_STRING")

	// Check if all required environment variables are set
	requiredEnvVars := []string{
		"SRC_DB_PASSWORD", "SRC_DB_URL_NAME", "SRC_DB_NAME", "SRC_DB_USERNAME",
		"DST_DB_PASSWORD", "DST_DB_URL_NAME", "DST_DB_NAME", "DST_DB_USERNAME",
	}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Required environment variable %s is not set", envVar)
		}
	}

	// use the environment variables
	// sourceDbConnString := fmt.Sprintf("host=%s.postgres.database.azure.com user=%s password=%s dbname=%s port=5432 sslmode=require", srcDbUrlName, srcDbUsername, srcDbPassword, srcDbName)
	sourceDb, err := gorm.Open(postgres.Open(srcDbConnString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to source PostgreSQL: %v", err)
	}
	log.Println("Connected to source PostgreSQL database successfully!")

	// Connect to Azure SQL database
	// destinationDbConnString := fmt.Sprintf("server=%s.database.windows.net;user id=%s;password=%s;port=1433;database=%s;", dstDbUrlName, dstDbUsername, dstDbPassword, dstDbName)
	//targetDb, err := gorm.Open(sqlserver.Open(destinationDbConnString), &gorm.Config{})
	targetDb, err := gorm.Open(postgres.Open(dstDbConnString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Azure SQL: %v", err)
	}
	log.Printf("targetDb: %v", targetDb)
	log.Printf("sourceDb: %v", sourceDb)

	var plantRecords = getPlantRecords(sourceDb)
	fmt.Println("Found ", len(plantRecords), "records in source PostgreSQL.")
	var imageRecords = getImageRecords(sourceDb)
	fmt.Println("Found ", len(imageRecords), "records in source PostgreSQL.")

	// make a map of the image id to sha256 hash of the image data so we can look up a sha256 and find the image id
	var imageHashMap = make(map[int]string)
	for _, image := range imageRecords {
		sum := sha256.Sum256(image.Data)
		// fmt.Printf("imageId: %d %x", image.ID, sum)
		var sumString = fmt.Sprintf("%x", sum) // convert [32]byte to string
		imageHashMap[int(image.ID)] = sumString
	}

	// make a map of plant name to plant image id
	var plantImageMap = make(map[string]int)
	for _, plant := range plantRecords {
		plantImageMap[plant.Name] = plant.ImageId
		expectedSha := imageHashMap[plant.ImageId]
		log.Printf("Plant: %s ImageId: %d, expectedSha %x", plant.Name, plant.ImageId, expectedSha)

	}

	// dropTables(targetDb)

	// if *dropOnly {
	// 	fmt.Println("Drop flag set, exiting after dropping tables")
	// 	return
	// }

	// var _ = migratePlants(sourceDb, targetDb)
	// // migrateComments(sourceDb, targetDb, plantIdMap)
	// // migratePlantLogs(sourceDb, targetDb, plantIdMap)
	// log.Printf("Migrating plant images...")
	// var imageIdMap = migrateImages(sourceDb, targetDb)

	// // now update the plant records with the new imageId
	// var plants []models.PlantModel
	// if err := targetDb.Find(&plants).Error; err != nil {
	// 	log.Fatalf("Failed to fetch records from PostgreSQL: %v", err)
	// }
	// fmt.Println("Found ", len(plants), "records in PostgreSQL.")
	// for _, plant := range plants {
	// 	var newImageId = imageIdMap[plant.ImageId]
	// 	fmt.Println("Updating plant old imageId: ", plant.ImageId, " Image ID: ", newImageId)
	// 	plant.ImageId = newImageId
	// 	if err := targetDb.Save(&plant).Error; err != nil {
	// 		log.Printf("Failed to upsert record ID %d: %v", plant.ID, err)
	// 	}
	// }
	// targetDb.Commit()

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
	fmt.Println("Found ", len(records), "records in source PostgreSQL.")

	var numUpdated = 0

	// Synchronize data to Azure SQL
	for _, record := range records {
		var oldId = record.ID
		if err := targetDb.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&record).Error; err != nil {
			log.Fatalf("Failed to upsert record ID %d: %v", record.ID, err)
		}
		numUpdated++
		var newId = record.ID
		plantIdMap[oldId] = newId
	}
	fmt.Printf("Synchronization complete, migrated %d plants out of %d!\n", len(records), numUpdated)

	// print the plantIdMap
	fmt.Println("Plant ID Map:")
	for oldId, newId := range plantIdMap {
		fmt.Printf("Old ID: %d -> New ID: %d\n", oldId, newId)
	}
	return plantIdMap
}
