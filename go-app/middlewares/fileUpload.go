package middlewares

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/LidoHon/recipes-server/libs"
	"github.com/LidoHon/recipes-server/models"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

const maxFileSize = 5 * 1024 * 1024 // 5MB

func sanitizeFileName(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName)) // Remove existing extension
}

func handleImageUpload(image models.ImageInput) (string, string) {
	// Check if the file size exceeds the limit
	if len(image.Base64String) > maxFileSize {
		return "", "File size exceeds the maximum limit of 5MB"
	}

	// Check if the file type is allowed
	if !allowedMimeTypes[image.Type] {
		return "", "Invalid file type"
	}

	// Decode base64 string to []byte
	imageData, err := base64.StdEncoding.DecodeString(image.Base64String)
	if err != nil {
		fmt.Println("decoding image error:", err.Error())
		return "", "Failed to decode image"
	}

	// Sanitize file name
	sanitizedFileName := sanitizeFileName(image.Name)

	// Initialize Cloudinary client
	cld, err := libs.SetupCloudinary()
	if err != nil {
		fmt.Println("setting up cloudinary error:", err.Error())
		return "", "Failed to initialize Cloudinary"
	}

	// Wrap the image data in a bytes.Reader
	imageReader := bytes.NewReader(imageData)

	// Upload image to Cloudinary
	uploadResult, err := cld.Upload.Upload(context.Background(), imageReader, uploader.UploadParams{PublicID: sanitizedFileName})
	if err != nil {
		fmt.Println("uploading image error:", err.Error())
		return "", "Failed to upload image"
	}

	// Return the URL of the uploaded image
	return uploadResult.SecureURL, ""
}

func ImageUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the body data
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("body parsing error:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}

		// Restore the body so that the controller can read it again
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Extract the image data only
		var requestBody struct {
			Input struct {
				Image  *models.ImageInput  `json:"image"`
				Images []models.ImageInput `json:"images"`
			} `json:"input"`
		}

		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			fmt.Println("error in unmarshal the body:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image data"})
			c.Abort()
			return
		}

		if requestBody.Input.Image != nil {
			fmt.Println("Single image upload detected")
			imageUrl, err := handleImageUpload(*requestBody.Input.Image)
			if err != "" {
				fmt.Println("error in handling single image upload:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image", "detail": err})
				c.Abort()
				return
			}
			fmt.Println("Setting imageUrl in context:", imageUrl)
			c.Set("imageUrl", imageUrl)
		} else if len(requestBody.Input.Images) > 0 {
			fmt.Println("Multiple image uploads detected")
			var imageUrls []string
			for _, image := range requestBody.Input.Images {
				imageUrl, err := handleImageUpload(image)
				if err != "" {
					fmt.Println("error in multiple image upload:", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload images", "detail": err})
					c.Abort()
					return
				}
				imageUrls = append(imageUrls, imageUrl)
			}
			fmt.Println("Setting imageUrls in context:", imageUrls)
			c.Set("imageUrls", imageUrls)
		} else {
			fmt.Println("No image data found in request body")
		}
		c.Next()
	}
}