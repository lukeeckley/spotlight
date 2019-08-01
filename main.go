//go:generate goversioninfo
package main

import (
	"crypto/md5"
	"encoding/hex"
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
	"time"
	"strconv"

	"github.com/lxn/walk"
)

func main() {
	var imageCount int
	var imageCountMsg string

	// Taking this from https://github.com/lxn/walk/blob/master/examples/notifyicon/notifyicon.go
	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}
	// We load our icon from a file.
	icon, err := walk.Resources.Icon("pinkparty.ico")
	if err != nil {
		log.Fatal(err)
	}

	// Create the notify icon and make sure we clean it up on exit.
	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}
	defer ni.Dispose()

	// Set the icon and a tool tip text.
	if err := ni.SetIcon(icon); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetToolTip("Processing Spotlight Photos..."); err != nil {
		log.Fatal(err)
	}

	// When the left mouse button is pressed, bring up our balloon.
	/* ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		}

		if err := ni.ShowCustom(
			"Walk NotifyIcon Example",
			"There are multiple ShowX methods sporting different icons.",
			icon); err != nil {

			log.Fatal(err)
		}
	})*/

	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}

	// The notify icon is hidden initially, so we have to make it visible.
	if err := ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	spotlightfolder := usr.HomeDir + "\\AppData\\Local\\Packages\\Microsoft.Windows.ContentDeliveryManager_cw5n1h2txyewy\\LocalState\\Assets\\"
	outputfolder := usr.HomeDir + "\\Desktop\\Spotlight\\"
	if _, err := os.Stat(outputfolder); os.IsNotExist(err) {
		os.Mkdir(outputfolder, 0700)
	}
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
					err := copyFile(filePath, outputfolder+hash+".jpg")
					if err != nil {
						log.Fatal(err)
					}
					imageCount++
				}
			}
		}
	}

	if (imageCount == 0) {
		imageCountMsg = "Didn't copy any spotlight photos"
	} else {
		imageCountMsg = "Copied " + strconv.Itoa(imageCount) + " new spotlight photos to your desktop"
	}

	if err := ni.ShowCustom("Spotlight", imageCountMsg, icon); err != nil {
		log.Fatal(err)
	}

	// If we don't sleep the program closes and we don't see the notification telling us how many photos were copied
	time.Sleep(5 * time.Second)
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
	}
	return false
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
