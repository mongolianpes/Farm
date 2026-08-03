package images

import "os"

const (
	pathToImages       = "/images/"
	pathToDefaultImage = "/images/d.webp"
)

var imagesServiceExternalConnections = os.Getenv("IMAGES_SERVICE_EXTERNAL_CONNECTIONS")

func MakeCurrentPathToImage(imageName string) string {
	if imageName == "" {
		return imagesServiceExternalConnections + pathToDefaultImage
	}
	return imagesServiceExternalConnections + pathToImages + imageName
}
