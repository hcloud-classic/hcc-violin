package uuidgen

import (
	"hcloud-violin/logger"

	uuid "github.com/nu7hatch/gouuid"
)

// Uuidgen : Generate uuid
func Uuidgen() (string, error) {
	out, err := uuid.NewV4()
	if err != nil {
		logger.Log.Println(err)
		return "", err
	}

	return out.String(), nil
}
