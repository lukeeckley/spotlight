package main

import (
	"crypto/md5"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"encoding/hex"
)

func main() {
	usr, _ := user.Current()
	spotlightfolder := usr.HomeDir + "\\AppData\\Local\\Packages\\Microsoft.Windows.ContentDeliveryManager_cw5n1h2txyewy\\LocalState\\Assets\\"
	outputfolder := usr.HomeDir + "\\Desktop\\MyPhotos\\"
	files, err := ioutil.ReadDir(spotlightfolder)
	if err != nil {
		log.Fatal(err)
	}

	// Get slice of all the hashes in the output folder
	var hashes = getHashesFromWallpapers(outputfolder)

	for _, f := range files {
		filePath := spotlightfolder + f.Name()
		////fmt.Printf(filePath)
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}
		file.Seek(0, 0)
		contentType := http.DetectContentType(buffer)

		if contentType == "image/jpeg" {
			if isWallpaper(filePath) {
				h := md5.New()
				if _, err := io.Copy(h, file); err != nil {
					log.Fatal(err)
				}
				hash := hex.EncodeToString(h.Sum(nil))
				if !hashInHashes(hash, hashes) {
					err := copyFile(filePath, outputfolder + f.Name() + ".jpg")
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}
}

func getHashesFromWallpapers(folder string) []string {
	var hashes []string
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		filePath := folder + f.Name()
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		h := md5.New()
		if _, err := io.Copy(h, file); err != nil {
			log.Fatal(err)
		}
		hashes = append(hashes, hex.EncodeToString(h.Sum(nil)))
	}
	return hashes
}

func hashInHashes(hash string, hashes []string) bool {
	for _, v := range hashes {
		if v == hash {
			return true
		}
  }
	return false
}

// Only determines if image file is a wallpaper by determining if the image is wider than 1900 pixels
// Default Windows Spotlight Wallpaper size is 1920
func isWallpaper(imagePath string) bool {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}

	if image.Width > 1900 {
		return true
	} else {
		return false
	}
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}
