package cmd

import (
	"fmt"
	"lassebomh/minecraft-img2text/pkg"
	"time"

	"github.com/kbinani/screenshot"
	"gocv.io/x/gocv"
)

func ReadSignLoop(display int) {
	locationSigns, err := pkg.AutoSavedLocationSigns("./out/signs.json")

	if err != nil {
		panic(err)
	}

	for {
		screenshot, _ := screenshot.CaptureDisplay(display)
		img, _ := gocv.ImageToMatRGB(screenshot)

		signsLines, signsAreas := pkg.GetSignsLines(img)

		for i := range signsLines {
			lines := signsLines[i]
			area := signsAreas[i]

			fmt.Println(area, lines)

			sign := pkg.Sign{
				Lines:     lines,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Area:      area,
			}

			locationSigns.AddOrUpdateSign("global", sign)

		}

		img.Close()
	}
}
