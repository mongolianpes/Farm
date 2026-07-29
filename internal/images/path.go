package images

const (
	ImagesServiceExternalConnections = "http://localhost:8080"
	PathToImages                     = ImagesServiceExternalConnections + "/images/"
	PathToDefaultImage               = ImagesServiceExternalConnections + "/images/d.webp"
	DefaultImage                     = "d"
)

func MakeCurrentPathToImage(imageName string) string {
	if imageName == "" {
		return PathToDefaultImage
	}
	return PathToImages + imageName
}
