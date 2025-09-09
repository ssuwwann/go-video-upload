package fsutil

import "path/filepath"

func OriginalsDir(root, id string) string {
	return filepath.Join(root, "originals", id)
}

func OutputsDir(root, id string) string {
	return filepath.Join(root, "outputs", id)
}

func ThumbnailsDir(root, id string) string {
	return filepath.Join(root, "thumbnails", id)
}

func MetadataDir(root string) string {
	return filepath.Join(root, "metadata")
}
