package validations

import (
	"errors"
	"mime/multipart"
	"strconv"
	"sequAcc/models"
)

// ValidateShiningStarInput validates the input for creating a shining star
func ValidateShiningStarInput(userIDStr, achievementType, description string, file *multipart.FileHeader) (uint64, models.AchievementType, error) {
	if userIDStr == "" {
		return 0, "", errors.New("user_id is required")
	}
	if description == "" {
		return 0, "", errors.New("description is required")
	}

	validType := false
	parsedType := models.AchievementType(achievementType)
	switch parsedType {
	case models.TypeTaskMaster, models.TypeConsistencyStar, models.TypeSpeedPerformer, models.TypeDeadlineChampion, models.TypeProductivityPro:
		validType = true
	}

	if !validType {
		return 0, "", errors.New("invalid or missing achievement type")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, "", errors.New("invalid user_id format")
	}
	if file == nil {
		return 0, "", errors.New("certificate file is required")
	}
	if file.Size > 10*1024*1024 {
		return 0, "", errors.New("file size exceeds 10MB limit")
	}
	return userID, parsedType, nil
}
