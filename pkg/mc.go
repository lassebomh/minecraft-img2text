package pkg

import "github.com/go-vgo/robotgo"

func EnterCommand(s string) {
	robotgo.KeyTap("backspace")

	for i := range s {
		robotgo.KeyTap(string(robotgo.CharCodeAt(s, i)))
	}
	robotgo.KeyTap("enter")
}

func init() {
	robotgo.KeySleep = 30
}
