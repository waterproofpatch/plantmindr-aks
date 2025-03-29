// encapsulate database/storage operations
package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Helper function to check the database connection and ping it
func checkDBConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("Failed to get database connection:", err)
		return err
	}

	// Retry the connection 5 times with a 2-second interval
	for i := 0; i < 5; i++ {
		if err := sqlDB.Ping(); err != nil {
			fmt.Printf("Database is not available (attempt %d): %v\n", i+1, err)
			time.Sleep(2 * time.Second)
		} else {
			fmt.Println("Database connection successful")
			return nil
		}
	}

	// Return the last error if all retries fail
	fmt.Println("Database connection failed after 5 attempts")
	return fmt.Errorf("database connection failed after 5 attempts")
}

func GetPlants(db *gorm.DB, email string, plants *[]PlantModel) error {
	// Use the helper function to check the database connection
	if err := checkDBConnection(db); err != nil {
		return err
	}

	db.Where("email = ? OR is_public = ?", email, true).Preload("Logs").Preload("Comments").Find(&plants)
	if db.Error != nil {
		fmt.Println("Had an error getting plants:", db.Error)
		return db.Error
	}
	fmt.Println("done getting plants")
	return nil
}
