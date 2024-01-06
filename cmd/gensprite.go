package cmd

import (
	"lassebomh/minecraft-img2text/pkg"

	"gocv.io/x/gocv"
)

func GenSprite() {
	mat := pkg.AsciiSpriteImg()
	defer mat.Close()

	gocv.IMWrite("./resourcepacks/MachineReader/assets/minecraft/textures/font/ascii.png", mat)
}
