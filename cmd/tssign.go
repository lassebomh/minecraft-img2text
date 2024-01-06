package cmd

import (
	"fmt"
	"lasse/minecraft-sign-reader/pkg"
	"strconv"
	"strings"

	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
	hook "github.com/robotn/gohook"
	"gocv.io/x/gocv"
	"gorm.io/gorm"
)

var shopId string = "b2u"
var shopLastInt int = 64

func add(pause chan bool) {
	hook.Register(hook.KeyDown, []string{"tab", "t"}, func(e hook.Event) {
		pause <- true
		fmt.Println("Sending pause signal. Screenshot operations stopped.")

		robotgo.Sleep(1)

		shopLastInt += 1

		pkg.EnterCommand("/arm tp " + shopId + strconv.Itoa(shopLastInt))

		robotgo.MilliSleep(500)

		pause <- true
		fmt.Println("Sending resume signal. Resuming screenshot operations.")
	})

	s := hook.Start()
	<-hook.Process(s)
}

type InvalidItemError struct{}

func (m *InvalidItemError) Error() string {
	return "Invalid item error"
}

func ExtractItem(input string) (int, string, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) != 2 {
		return 0, "", &InvalidItemError{}
	}
	n, err := strconv.Atoi(parts[0])
	return n, parts[1], err
}

func TradeScanLoop(db *gorm.DB) {

	var _shopId string
	var _shopLastInt int

	fmt.Print("shop row: ")

	fmt.Scanln(&_shopId)

	fmt.Print("shop num: ")

	fmt.Scanln(&_shopLastInt)

	shopId = _shopId
	shopLastInt = _shopLastInt

	pause := make(chan bool, 1)

	go add(pause)

	for {
		select {
		case <-pause:
			<-pause
		default:
			screenshot, _ := screenshot.CaptureDisplay(0)
			img, _ := gocv.ImageToMatRGB(screenshot)

			signsLines, signsAreas := pkg.GetSignsLines(img)

			for i := range signsLines {
				lines := signsLines[i]
				area := signsAreas[i]

				if lines[0] != "[Trade]" || (lines[3] != "<Out Of Stock>" && lines[3] != "<Open>") {
					continue
				}

				productQuantity, product, err := ExtractItem(lines[1])

				if err != nil {
					fmt.Println(err)
					continue
				}

				costQuantity, cost, err := ExtractItem(lines[2])

				if err != nil {
					fmt.Println(err)
					continue
				}

				inStock := lines[3] == "<Open>"

				// t, err := trade.FindTradeAndUpdateOrCreate(db, area, productQuantity, product, costQuantity, cost, shopId+strconv.Itoa(shopLastInt), inStock)

				fmt.Println(db, area, productQuantity, product, costQuantity, cost, shopId+strconv.Itoa(shopLastInt), inStock)

				// if t != nil {
				// 	fmt.Println(t.ID, t.ShopID)
				// }

				// if err != nil {
				// 	fmt.Println(err)
				// }
			}

			img.Close()
		}
	}
}
