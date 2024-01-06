package pkg

import (
	"image"
	"image/color"
	"math"

	"gocv.io/x/gocv"
)

var validRuneStart = ' '
var validRuneEnd = '~'

func DisplayImage(window *gocv.Window, img gocv.Mat) {
	for {
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func bresenhamPoints(p1, p2 image.Point) []image.Point {
	dx := int(math.Abs(float64(p2.X) - float64(p1.X)))
	sx := -1
	if p1.X < p2.X {
		sx = 1
	}

	dy := -int(math.Abs(float64(p2.Y) - float64(p1.Y)))
	sy := -1
	if p1.Y < p2.Y {
		sy = 1
	}

	err := dx + dy

	points := []image.Point{}

	for {
		points = append(points, image.Point{p1.X, p1.Y})

		if p1.X == p2.X && p1.Y == p2.Y {
			break
		}

		e2 := 2 * err

		if e2 >= dy {
			if p1.X == p2.X {
				break
			}
			err = err + dy
			p1.X = p1.X + sx
		}

		if e2 <= dx {
			if p1.Y == p2.Y {
				break
			}
			err = err + dx
			p1.Y = p1.Y + sy
		}
	}

	return points
}

func min3(a, b, c float64) float64 {
	return math.Min(math.Min(a, b), c)
}

func hueToRGB(h float64) (uint8, uint8, uint8) {
	kr := math.Mod(5+h*6, 6)
	kg := math.Mod(3+h*6, 6)
	kb := math.Mod(1+h*6, 6)

	r := 1 - math.Max(min3(kr, 4-kr, 1), 0)
	g := 1 - math.Max(min3(kg, 4-kg, 1), 0)
	b := 1 - math.Max(min3(kb, 4-kb, 1), 0)

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

var ColorToRune map[color.RGBA]rune
var RuneToColor map[rune]color.RGBA

var tlColors = []color.RGBA{
	{187, 184, 7, 255},
	{125, 123, 5, 255},
}
var trColors = []color.RGBA{
	{7, 187, 88, 255},
	{5, 125, 59, 255},
}
var blColors = []color.RGBA{
	{10, 7, 187, 255},
	{7, 5, 125, 255},
}
var brColors = []color.RGBA{
	{187, 7, 103, 255},
	{125, 5, 69, 255},
}

func colorInRange(r uint8, g uint8, b uint8, colors []color.RGBA, threshold float64) bool {

	for i := range colors {
		color := colors[i]

		if math.Abs(float64(color.R)-float64(r)) < threshold && math.Abs(float64(color.G)-float64(g)) < threshold && math.Abs(float64(color.B)-float64(b)) < threshold {
			return true
		}
	}

	return false
}

func AsciiSpriteImg() gocv.Mat {

	charRows, charCols := 16, 16
	scale := 4

	img := gocv.NewMatWithSize(charRows*scale, charCols*scale, gocv.MatTypeCV8UC4)

	for row := 0; row < charRows; row++ {
		for col := 0; col < charCols; col++ {
			x, y := col*scale, row*scale
			r := rune(row*charRows + col)
			if r >= validRuneStart && r <= validRuneEnd {
				runeColor := RuneToColor[r]
				gocv.RectangleWithParams(&img, image.Rect(x, y, x, y+scale), runeColor, 0, gocv.Line4, 0)
				gocv.RectangleWithParams(&img, image.Rect(x+1, y, x+1, y+scale), color.RGBA{0, 0, 0, 1}, 0, gocv.Line4, 0)
			}
		}
	}

	return img
}

var validSignColors []color.RGBA = []color.RGBA{{0, 0, 0, 255}}

// var validSignColors []color.RGBA = []color.RGBA{{0, 0, 0, 255}, {255, 255, 255, 255}}

func labelNeighbors(labelMat gocv.Mat, nlabels int) ([]int, gocv.PointsVector) {
	height, width := labelMat.Rows(), labelMat.Cols()

	neighbors := gocv.NewPointsVector()
	areas := []int{}

	for l := 0; l < nlabels; l++ {
		neighbors.Append(gocv.NewPointVector())
		areas = append(areas, 0)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			label := labelMat.GetIntAt(y, x)
			areas[label] += 1
			if label == 0 {
				continue
			}
			if x > 0 && labelMat.GetIntAt(y, x-1) != label {
				neighbors.At(int(label)).Append(image.Point{x - 1, y})
			}
			if x < width-1 && labelMat.GetIntAt(y, x+1) != label {
				neighbors.At(int(label)).Append(image.Point{x + 1, y})
			}
			if y > 0 && labelMat.GetIntAt(y-1, x) != label {
				neighbors.At(int(label)).Append(image.Point{x, y - 1})
			}
			if y < height-1 && labelMat.GetIntAt(y+1, x) != label {
				neighbors.At(int(label)).Append(image.Point{x, y + 1})
			}
		}
	}

	return areas, neighbors
}

func getSigns(img gocv.Mat, bgr []gocv.Mat) ([][4]image.Point, []int) {

	corners := [][4]image.Point{}
	outAreas := []int{}

	binaryImg := gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8U)
	defer binaryImg.Close()

	binaryImg.SetTo(gocv.Scalar{Val1: 0})

	for _, col := range validSignColors {
		mask := gocv.NewMat()
		defer mask.Close()

		targetScalar := gocv.NewScalar(float64(col.B), float64(col.G), float64(col.R), float64(col.A))
		gocv.InRangeWithScalar(img, targetScalar, targetScalar, &mask)
		gocv.BitwiseOr(binaryImg, mask, &binaryImg)
	}

	labels := gocv.NewMat()
	defer labels.Close()

	nlabels := gocv.ConnectedComponents(binaryImg, &labels)

	areas, points := labelNeighbors(labels, nlabels)
	defer points.Close()

	for label := 1; label < nlabels; label++ {
		if areas[label] < 1500 {
			continue
		}

		// fmt.Println(areas[label])

		labelPoints := points.At(label).ToPoints()

		trX, trY, trN := 0, 0, 0
		tlX, tlY, tlN := 0, 0, 0
		brX, brY, brN := 0, 0, 0
		blX, blY, blN := 0, 0, 0

		for i := range labelPoints {
			point := labelPoints[i]
			r, g, b := bgr[2].GetUCharAt(point.Y, point.X), bgr[1].GetUCharAt(point.Y, point.X), bgr[0].GetUCharAt(point.Y, point.X)

			if colorInRange(r, g, b, trColors, 10) {
				if trN == 0 {
					trX, trY = point.X, point.Y
				} else {
					trX += point.X
					trY += point.Y
				}
				trN += 1
			}

			if colorInRange(r, g, b, tlColors, 10) {
				if tlN == 0 {
					tlX, tlY = point.X, point.Y
				} else {
					tlX += point.X
					tlY += point.Y
				}
				tlN += 1
			}

			if colorInRange(r, g, b, brColors, 10) {
				if brN == 0 {
					brX, brY = point.X, point.Y
				} else {
					brX += point.X
					brY += point.Y
				}
				brN += 1
			}

			if colorInRange(r, g, b, blColors, 10) {
				if blN == 0 {
					blX, blY = point.X, point.Y
				} else {
					blX += point.X
					blY += point.Y
				}
				blN += 1
			}
		}

		if trN == 0 || tlN == 0 || brN == 0 || blN == 0 {
			continue
		}

		tl := image.Point{tlX / tlN, tlY / tlN}
		tr := image.Point{trX / trN, trY / trN}
		bl := image.Point{blX / blN, blY / blN}
		br := image.Point{brX / brN, brY / brN}

		t := math.Sqrt(math.Pow(float64(tl.X)-float64(tr.X), 2) + math.Pow(float64(tl.Y)-float64(tr.Y), 2))
		b := math.Sqrt(math.Pow(float64(bl.X)-float64(br.X), 2) + math.Pow(float64(bl.Y)-float64(br.Y), 2))
		l := math.Sqrt(math.Pow(float64(tl.X)-float64(bl.X), 2) + math.Pow(float64(tl.Y)-float64(bl.Y), 2))
		r := math.Sqrt(math.Pow(float64(tr.X)-float64(br.X), 2) + math.Pow(float64(tr.Y)-float64(br.Y), 2))

		ratio := (t + b) / (2 * (l + r))

		// fmt.Println(float64(areas[label])*ratio, ratio)

		if float64(areas[label])*ratio < 1500 || ratio > 2.2 {
			continue
		}

		corners = append(corners, [4]image.Point{tl, tr, bl, br})
		outAreas = append(outAreas, areas[label])
	}

	return corners, outAreas
}

func interpolatePoints(a, b image.Point, frac float64) image.Point {
	return image.Point{
		int(float64(b.X-a.X)*frac + float64(a.X)),
		int(float64(b.Y-a.Y)*frac + float64(a.Y)),
	}
}

func readSign(img *gocv.Mat, bgr []gocv.Mat, corners [4]image.Point) (bool, [4]string) {

	tl, tr, bl, br := corners[0], corners[1], corners[2], corners[3]

	textLines := [4]string{"", "", "", ""}

	for lineI := 0; lineI < 4; lineI++ {
		frac := 0.2 + float64(lineI)*0.215
		points := bresenhamPoints(interpolatePoints(tl, bl, frac), interpolatePoints(tr, br, frac))

		lastChar := rune(0)

		for i := range points {
			point := points[i]

			r := uint8(bgr[2].GetUCharAt(point.Y, point.X))
			g := uint8(bgr[1].GetUCharAt(point.Y, point.X))
			b := uint8(bgr[0].GetUCharAt(point.Y, point.X))

			if r == 0 && g == 0 && b == 0 {
				lastChar = 0
				continue
			}

			color := color.RGBA{r, g, b, 255}
			char, ok := ColorToRune[color]

			if !ok {
				return false, textLines
			}

			if lastChar == 0 || lastChar != char {
				textLines[lineI] += string(char)
			}

			lastChar = char
		}
	}

	return true, textLines
}

func GetSignsLines(img gocv.Mat) ([][4]string, []int) {
	signsLines := [][4]string{}
	signsAreas := []int{}

	bgr := gocv.Split(img)

	signsCorners, areas := getSigns(img, bgr)

	for signI := range signsCorners {
		ok, lines := readSign(&img, bgr, signsCorners[signI])
		if ok {
			signsLines = append(signsLines, lines)
			signsAreas = append(signsAreas, areas[signI])
		}
	}

	bgr[0].Close()
	bgr[1].Close()
	bgr[2].Close()

	return signsLines, signsAreas
}

func ReadMouseText(img gocv.Mat, x int, y int, cx int, cy int) string {
	bgr := gocv.Split(img)

	defer bgr[0].Close()
	defer bgr[1].Close()
	defer bgr[2].Close()

	line := ""

	if x >= img.Cols() {
		x -= img.Cols()
	}

	if x < 0 || y < 0 || x >= img.Cols() || y >= img.Rows() {
		return ""
	}

	_, ok := ColorToRune[color.RGBA{uint8(bgr[2].GetUCharAt(y, x+cx)), uint8(bgr[1].GetUCharAt(y, x+cx)), uint8(bgr[0].GetUCharAt(y, x+cx)), 255}]

	if !ok {
		return ""
	}

	lastChar := rune(0)

	for x < img.Cols() && cx < 330 {
		cx++

		if x+cx < 0 || y < 0 || x+cx >= img.Cols() || y >= img.Rows() {
			continue
		}

		r := uint8(bgr[2].GetUCharAt(y+cy, x+cx))
		g := uint8(bgr[1].GetUCharAt(y+cy, x+cx))
		b := uint8(bgr[0].GetUCharAt(y+cy, x+cx))

		color := color.RGBA{r, g, b, 255}
		char, ok := ColorToRune[color]

		if !ok {
			lastChar = 0
			continue
		}

		if lastChar == 0 || lastChar != char {
			line += string(char)
		}

		lastChar = char

	}

	return line
}

func init() {

	RuneToColor = make(map[rune]color.RGBA)
	ColorToRune = make(map[color.RGBA]rune)

	for i := validRuneStart; i < validRuneEnd+1; i++ {

		r, g, b := hueToRGB(float64(i-validRuneStart) / float64((validRuneEnd+1)-validRuneStart))

		color := color.RGBA{r, g, b, 255}
		ColorToRune[color] = i
		RuneToColor[i] = color

		validSignColors = append(validSignColors, color)
	}
}
