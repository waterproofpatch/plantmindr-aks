package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/waterproofpatch/go_authentication/authentication"
	auth_types "github.com/waterproofpatch/go_authentication/types"

	"github.com/gorilla/mux"
)

func images(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageId, hasImageId := vars["id"]
	db := authentication.GetDb()
	fmt.Printf("imageId=%s\n", imageId)

	// frontend can request this when they get a plant with an imageId nonzero
	switch r.Method {
	case "GET":
		if hasImageId {
			imageIdNo, err := strconv.Atoi(imageId)
			if err != nil {
				WriteResponse(w, "Invalid image ID", http.StatusBadRequest, Generic)
				return
			}
			fmt.Printf("Handling request for imageId=%d\n", imageIdNo)

			var img ImageModel
			img.ID = uint(imageIdNo)
			if err := db.First(&img, img.ID).Error; err != nil {
				WriteResponse(w, "Failed loading image", http.StatusBadRequest, Generic)
				return
			}
			w.Header().Set("Content-WateringFrequency", "image/jpeg")
			w.Write(img.Data)
			return
		} else {
			WriteResponse(w, "Must supply image ID!", http.StatusBadRequest, Generic)
		}
	}
}

// get the version information from the environment for displaying to the frontend
func version(w http.ResponseWriter, r *http.Request) {
	// see docker-compose.dev
	version := os.Getenv("SITE_TIMESTAMP")
	isDebug := os.Getenv("IS_DEBUG")

	dict := map[string]string{
		"version": version,
		"isDebug": isDebug,
	}

	err := json.NewEncoder(w).Encode(dict)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func plants(w http.ResponseWriter, r *http.Request, claims *auth_types.JWTData) {
	db := authentication.GetDb()
	var plants []PlantModel
	var plant PlantModel
	vars := mux.Vars(r)
	id, hasPlantId := vars["id"]

	switch r.Method {
	case "GET":
		if hasPlantId {
			result := db.Where("id = ?", id).Preload("Logs").Preload("Comments").Find(&plant)
			fmt.Printf("%d record(s) found\n", result.RowsAffected)
			json.NewEncoder(w).Encode(plant)
			return
		}
	case "DELETE":
		if claims == nil {
			WriteResponse(w, "Must be logged in to delete plants.", http.StatusUnauthorized, Generic)
			return
		}
		if !hasPlantId {
			WriteResponse(w, "Must provide id!", http.StatusBadRequest, Generic)
			break
		}
		db.Find(&plant, id)
		if claims != nil && plant.Email != claims.Email {
			fmt.Printf("User %s tried deleting plant belonging to %s\n", claims.Email, plant.Email)
			WriteResponse(w, "This isn't your plant!", http.StatusBadRequest, Generic)
			return
		}
		fmt.Printf("Deleting plant imageId=%d\n", plant.ImageId)
		db.Delete(&ImageModel{}, plant.ImageId)
		fmt.Printf("Deleting plant id=%d\n", plant.ID)
		db.Delete(&PlantModel{}, id)
		break
	case "POST":
		if claims == nil {
			WriteResponse(w, "Must be logged in to add plants.", http.StatusUnauthorized, Generic)
			return
		}
		imageId := ImageUploadHandler(w, r)
		if imageId == 0 {
			fmt.Println("Upload did not contain an image.")
		}
		if imageId < 0 {
			fmt.Println("Bailing early, critical error handling image.")
			return
		}
		json.Unmarshal([]byte(r.FormValue("plant")), &plant)
		plant.ImageId = imageId
		plant.Email = claims.Email
		plant.Username = claims.Username

		fmt.Printf("Adding plant as: %v", plant)
		err := AddPlant(db, &plant)
		if err != nil {
			WriteResponse(w, err.Error(), http.StatusBadRequest, Generic)
			return
		}
		break
	case "PUT":
		if claims == nil {
			WriteResponse(w, "Must be logged in to edit plants.", http.StatusUnauthorized, Generic)
			return
		}
		// make a new plant based on form values
		err := json.Unmarshal([]byte(r.FormValue("plant")), &plant)
		if err != nil {
			WriteResponse(w, "Invalid plant ID", http.StatusBadRequest, Generic)
			return
		}

		// get the existing plant so we can obtain its old imageId
		var existingPlant PlantModel
		db.First(&existingPlant, plant.ID)
		if existingPlant.Email != claims.Email {
			fmt.Printf("User %s tried editing plant belonging to %s\n", claims.Email, existingPlant.Email)
			WriteResponse(w, "This isn't your plant!", http.StatusBadRequest, Generic)
			return
		}

		// conditionally upload a new image. An imageId of 0 means no image provided
		imageId := ImageUploadHandler(w, r)
		isNewImage := false
		if imageId == 0 {
			fmt.Println("Upload did not contain an image.")
			plant.ImageId = existingPlant.ImageId
		} else if imageId < 0 {
			fmt.Println("Critical error updating image. Bailing early")
			return
		} else {
			// only update imageId if the user changed the image. Otherwise,
			// they may have only updated non-image stuff. If they had the
			// default image before, they'll still have it. If they didn't update
			// their image, then their plant has the old imageId.
			plant.ImageId = imageId
			isNewImage = true

		}
		plant.Username = existingPlant.Username
		plant.Email = existingPlant.Email
		fmt.Printf("Updating plant id=%d to: %s", plant.ID, plant)

		err = UpdatePlant(db, &plant, isNewImage)
		if err != nil {
			WriteResponse(w, err.Error(), http.StatusBadRequest, Generic)
			return
		}
	}
	if claims != nil {
		// db.Where("email = ? OR is_public = ?", claims.Email, true).Preload("Logs").Preload("Comments").Find(&plants)
		err := GetPlants(db, claims.Email, &plants)
		if err != nil {
			WriteResponse(w, "Failed to get plants", http.StatusBadRequest, Generic)
			return
		}

		for i := range plants {
			for j := range plants[i].Comments {
				var count int64
				// if the plant is ours and we have not yet seen this comment...
				// TODO is this even necessary
				if !plants[i].Comments[j].Viewed && claims.Email == plants[i].Email {
					count += 1
				}
			}
		}
	} else {
		db.Where("is_public = ?", true).Preload("Logs").Preload("Comments").Find(&plants)
	}
	fmt.Printf("Encoding %d plants in response\n", len(plants))
	json.NewEncoder(w).Encode(plants)
}

func comments(w http.ResponseWriter, r *http.Request, claims *auth_types.JWTData) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	db := authentication.GetDb()
	plantId := "0"

	switch r.Method {
	case "GET":
		plantId = r.URL.Query().Get("plantId")
		break

	case "DELETE":
		commentId, hasCommentId := vars["id"]
		if !hasCommentId {
			WriteResponse(w, "Invalid commentId ID", http.StatusBadRequest, Generic)
			return
		}
		if claims == nil {
			WriteResponse(w, "Must be logged in to delete comments.", http.StatusUnauthorized, Generic)
			return
		}
		var comment CommentModel
		var plant PlantModel
		db.Where("id = ?", commentId).First(&comment)
		db.Where("id = ?", comment.PlantID).First(&plant)

		// users should be able to delete comments by others for their plant
		if plant.Email != claims.Email && comment.Email != claims.Email {
			WriteResponse(w, "This isn't your comment, nor a comment on your plant!", http.StatusBadRequest, Generic)
			return
		}

		db.Delete(&comment)

		plantId = strconv.FormatUint(uint64(plant.ID), 10)
	case "POST":
		if claims == nil {
			WriteResponse(w, "Must be logged in to post comments.", http.StatusUnauthorized, Generic)
			return
		}

		// Declare a new Comment struct.
		var comment CommentModel
		var plant PlantModel

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&comment)
		if err != nil {
			fmt.Printf("Error decoding comment: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("comment received: %v for plantId=%d", comment, comment.PlantID)

		// make sure the plant is public
		db.Where("id = ?", comment.PlantID).First(&plant)
		if !plant.IsPublic && plant.Email != claims.Email {
			WriteResponse(w, "This plant is not public and also not yours, you cannot comment on it!", http.StatusBadRequest, Generic)
			return
		}

		AddComment(db, comment.Content, claims.Email, claims.Username, comment.PlantID)
		plantId = strconv.FormatUint(uint64(plant.ID), 10)
	}
	fmt.Printf("getting comments for plantId=%v", plantId)
	var comments []CommentModel
	var plant PlantModel
	db.Where("plant_id = ?", plantId).Find(&comments)
	db.Where("id = ?", plantId).Find(&plant)

	// update all comments here as viewed if the plant the comments are for is the owners plant
	if claims != nil {
		if plant.Email == claims.Email {
			fmt.Println("Owner of the plant is viewing comments, marking as viewed...")
			for i := range comments {
				comments[i].Viewed = true
				db.Save(&comments[i])
			}
		}
	}
	json.NewEncoder(w).Encode(comments)
}

func InitViews(router *mux.Router) {
	router.HandleFunc("/api/comments", authentication.VerifiedOnly(comments, true)).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
	router.HandleFunc("/api/comments/{id:[0-9]+}", authentication.VerifiedOnly(comments, true)).Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
	router.HandleFunc("/api/plants", authentication.VerifiedOnly(plants, true)).Methods("GET", "POST", "PUT", "OPTIONS")
	router.HandleFunc("/api/plants/{id:[0-9]+}", authentication.VerifiedOnly(plants, true)).Methods("GET", "POST", "DELETE", "PUT", "OPTIONS")
	router.HandleFunc("/api/images/{id:[0-9]+}", images).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/version", version).Methods("GET", "OPTIONS")
}
