package images

import "os"

const (
	imagesServiceAddr  = "http://localhost:8080"
	pathToImages       = imagesServiceAddr + "/images/"
	pathToDefaultImage = imagesServiceAddr + "/images/d.webp"
)

var imagesServiceExternalConnections = os.Getenv("IMAGES_SERVICE_EXTERNAL_CONNECTIONS")

func MakeCurrentPathToImage(imageName string) string {
	if imageName == "" {
		return imagesServiceExternalConnections + pathToDefaultImage
	}
	return imagesServiceExternalConnections + pathToImages + imageName
}
