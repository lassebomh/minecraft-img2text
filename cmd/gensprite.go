package cmd

import (
	"lassebomh/minecraft-img2text/pkg"

	"gocv.io/x/gocv"
)

func GenSprite() {
	mat := pkg.AsciiSpriteImg()
	defer mat.Close()

	gocv.IMWrite("./resourcepacks/Machine Reader/assets/minecraft/textures/font/ascii.png", mat)
}
