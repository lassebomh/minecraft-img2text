package cmd

import (
	"fmt"
	"io/ioutil"
	"lassebomh/minecraft-img2text/pkg"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/kbinani/screenshot"
	hook "github.com/robotn/gohook"
	"gocv.io/x/gocv"
)

func processString(input string) error {

	trimmedString := strings.TrimSpace(input)

	if len(trimmedString) > 0 && unicode.IsLower(rune(trimmedString[0])) {
		return nil
	}

	set := loadSetFromFile("strings_set.txt")
	if set == nil {
		set = make(map[string]bool)
	}

	if !set[trimmedString] {

		set[trimmedString] = true

		fmt.Println(trimmedString)
		return appendStringToFile("strings_set.txt", trimmedString)
	}

	return nil
}

func loadSetFromFile(filename string) map[string]bool {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]bool)
		}
		fmt.Println("Error reading file:", err)
		return nil
	}

	set := make(map[string]bool)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if len(line) > 0 {
			set[line] = true
		}
	}
	return set
}

func appendStringToFile(filename, text string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(text + "\n"); err != nil {
		return err
	}
	return nil
}

type AtomicCoords struct {
	x int64
	y int64
}

var coords AtomicCoords

func CursorScanLoop(display int, offsetx int, offsety int) {
	atomic.StoreInt64(&coords.x, 0)
	atomic.StoreInt64(&coords.y, 0)

	hook.Register(hook.MouseMove, []string{}, func(e hook.Event) {
		atomic.StoreInt64(&coords.x, int64(e.X))
		atomic.StoreInt64(&coords.y, int64(e.Y))
	})

	go func() {
		for {
			x := atomic.LoadInt64(&coords.x)
			y := atomic.LoadInt64(&coords.y)

			screenshot, _ := screenshot.CaptureDisplay(display)
			img, _ := gocv.ImageToMatRGB(screenshot)

			text := pkg.ReadMouseText(img, int(x), int(y), offsetx, offsety)

			if text != "" {
				processString(text)
			}

			img.Close()

			time.Sleep(time.Millisecond)
		}
	}()

	s := hook.Start()
	<-hook.Process(s)
}
