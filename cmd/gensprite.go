package cmd

import (
	"lasse/minecraft-sign-reader/pkg"

	"gocv.io/x/gocv"
)

func GenSprite() {
	mat := pkg.AsciiSpriteImg()
	defer mat.Close()

	gocv.IMWrite("./resourcepacks/Machine Reader/assets/minecraft/textures/font/ascii.png", mat)
}
