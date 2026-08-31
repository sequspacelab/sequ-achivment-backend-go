package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"sequAcc/database"
	"sequAcc/models"
	"sequAcc/validations"
)

// CreateShiningStar godoc
// @Summary      Create a new shining star achievement
// @Description  Create a new shining star achievement for a user with a certificate upload
// @Tags         achievements
// @Accept       multipart/form-data
// @Produce      json
// @Param        user_id formData int true "User ID"
// @Param        type formData string true "Achievement Type (e.g. Task Master, Consistency Star, Speed Performer, Deadline Champion, Productivity Pro, 1 trip around)"
// @Param        description formData string true "Description"
// @Param        certificate formData file false "Certificate File"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /achievements/shining-star [post]
func CreateShiningStar(c *gin.Context) {
	userIDStr := c.PostForm("user_id")
	if userIDStr == "" {
		userIDStr = c.PostForm("userId") // Fallback for camelCase
	}
	if userIDStr == "" {
		userIDStr = c.Query("user_id") // Fallback to query params
	}

	achievementTypeStr := c.PostForm("type")
	description := c.PostForm("description")

	// Fallback to JSON if the client mistakenly sends application/json instead of multipart/form-data
	if c.ContentType() == "application/json" {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if val, ok := body["user_id"]; ok && userIDStr == "" {
				userIDStr = fmt.Sprintf("%v", val)
			}
			if val, ok := body["userId"]; ok && userIDStr == "" {
				userIDStr = fmt.Sprintf("%v", val)
			}
			if val, ok := body["type"]; ok && achievementTypeStr == "" {
				achievementTypeStr = fmt.Sprintf("%v", val)
			}
			if val, ok := body["description"]; ok && description == "" {
				description = fmt.Sprintf("%v", val)
			}
		}
	}

	file, _ := c.FormFile("certificate")

	userID, achievementType, err := validations.ValidateShiningStarInput(userIDStr, achievementTypeStr, description, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin ID from context (set by JWT middleware)
	// Fallback to 0 if not present or cannot parse
	var adminID uint
	if adminIDVal, exists := c.Get("user_id"); exists {
		if id, ok := adminIDVal.(float64); ok {
			adminID = uint(id)
		}
	}

	var certificateURL string
	
	if file != nil {
		// Generate a unique filename
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
		
		// Open the uploaded file
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open certificate file"})
			return
		}
		defer f.Close()

		// Load AWS/S3 Config
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(os.Getenv("AWS_REGION")),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			)),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load s3 config"})
			return
		}

		// Create S3 client
		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
				o.BaseEndpoint = aws.String(endpoint)
				o.UsePathStyle = true // Common requirement for non-AWS S3 storage (like Hetzner, MinIO)
			}
		})

		bucket := os.Getenv("AWS_S3_BUCKET")
		if bucket == "" {
			bucket = "vertex" // Fallback bucket
		}

		// Upload to S3
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String("certificates/" + filename),
			Body:   f,
			ContentType: aws.String(file.Header.Get("Content-Type")), // Pass the correct content type
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload certificate to S3"})
			return
		}

		// Construct public URL
		if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
			certificateURL = fmt.Sprintf("%s/%s/certificates/%s", endpoint, bucket, filename)
		} else {
			certificateURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/certificates/%s", bucket, os.Getenv("AWS_REGION"), filename)
		}
	}

	achievement := models.ShiningStar{
		UserID:         uint(userID),
		AdminID:        adminID,
		Type:           achievementType,
		Description:    description,
		CertificateURL: certificateURL,
	}

	if err := database.DB.Create(&achievement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create achievement record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Shining star created successfully",
		"data":    achievement,
	})
}

// GetUserShiningStars godoc
// @Summary      Get user's shining stars
// @Description  Get a list of shining stars for a specific user
// @Tags         achievements
// @Produce      json
// @Param        user_id path int true "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /achievements/shining-star/{user_id} [get]
func GetUserShiningStars(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var achievements []models.ShiningStar
	if err := database.DB.Where("user_id = ?", userID).Find(&achievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch achievements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": achievements,
	})
}

// GetAllShiningStars godoc
// @Summary      Get all shining stars
// @Description  Get a list of all shining stars across the system
// @Tags         achievements
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /achievements/shining-star [get]
func GetAllShiningStars(c *gin.Context) {
	var achievements []models.ShiningStar
	if err := database.DB.Find(&achievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch achievements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": achievements,
	})
}
