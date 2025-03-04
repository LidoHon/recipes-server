package libs

import (
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

func SetupCloudinary() (*cloudinary.Cloudinary, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUD_NAME"),
		os.Getenv("COULDINRAY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return nil, err
	}
	return cld, nil
}
