package cmd

import (
	"fmt"
	"lasse/minecraft-sign-reader/pkg"

	"github.com/kbinani/screenshot"
	"gocv.io/x/gocv"
	"gorm.io/gorm"
)

func ReadShopLoop(db *gorm.DB) {
	for {
		screenshot, _ := screenshot.CaptureDisplay(0)
		img, _ := gocv.ImageToMatRGB(screenshot)

		signsLines, signsAreas := pkg.GetSignsLines(img)

		for i := range signsLines {
			lines := signsLines[i]
			area := signsAreas[i]

			if lines[0] != "Sold" {
				continue
			}

			_shopId := lines[1]
			_owner := lines[3]

			fmt.Println(area, _shopId, _owner)
		}

		img.Close()
	}
}
