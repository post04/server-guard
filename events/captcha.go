package events

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/post04/gomath/rational"
)

var (
	red    = colorr{r: 255, g: 50, b: 50, a: 255, name: "red", usage: 0, list: [][]int{}}
	orange = colorr{r: 255, g: 130, b: 21, a: 255, name: "orange", usage: 0, list: [][]int{}}
	yellow = colorr{r: 220, g: 220, b: 50, a: 255, name: "yellow", usage: 0, list: [][]int{}}
	green  = colorr{r: 50, g: 130, b: 50, a: 255, name: "green", usage: 0, list: [][]int{}}
	blue   = colorr{r: 80, g: 110, b: 220, a: 255, name: "blue", usage: 0, list: [][]int{}}
	purple = colorr{r: 170, g: 50, b: 190, a: 255, name: "purple", usage: 0, list: [][]int{}}
	white  = colorr{r: 255, g: 255, b: 255, a: 255, name: "white", usage: 0, list: [][]int{}}
	black  = colorr{r: 0, g: 0, b: 0, a: 255, name: "black", usage: 0, list: [][]int{}}
	brown  = colorr{r: 100, g: 50, b: 40, a: 255, name: "brown", usage: 0, list: [][]int{}}
	pink   = colorr{r: 230, g: 110, b: 210, a: 255, name: "pink", usage: 0, list: [][]int{}}
	gray   = colorr{r: 160, g: 160, b: 160, a: 255, name: "gray", usage: 0, list: [][]int{}}
)

type colorr struct {
	r, g, b, a, usage int
	name              string
	list              [][]int
}

func init() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

// NewCaptcha returns an io.Reader that contains the image warped
func (c *Config) NewCaptcha() (io.Reader, string, error) {
	emoji := c.Emojis[rand.Intn(len(c.Emojis))]
	imgfile, err := os.Open("emojis/" + emoji + ".png")
	if err != nil {
		return nil, "", err
	}
	defer imgfile.Close()
	imgCfg, _, err := image.DecodeConfig(imgfile)
	if err != nil {
		return nil, "", err
	}
	colors := []string{}
	width := imgCfg.Width
	height := imgCfg.Height

	imgfile.Seek(0, 0)

	img, _, err := image.Decode(imgfile)

	cimg := image.NewRGBA(img.Bounds())
	draw.Draw(cimg, img.Bounds(), img, image.Point{}, draw.Over)
	//finds usage of each color
	for x := 1; x <= width; x++ {
		for y := 1; y <= height; y++ {
			colorr, rgba := findColor(img.At(x, y).RGBA())
			colors = append(colors, colorr)
			switch colorr {
			case red.name:
				red.list = append(red.list, rgba)
				red.usage++
			case orange.name:
				orange.list = append(orange.list, rgba)
				orange.usage++
			case yellow.name:
				yellow.list = append(yellow.list, rgba)
				yellow.usage++
			case green.name:
				green.list = append(green.list, rgba)
				green.usage++
			case blue.name:
				blue.list = append(blue.list, rgba)
				blue.usage++
			case purple.name:
				purple.list = append(purple.list, rgba)
				purple.usage++
			case black.name:
				black.list = append(black.list, rgba)
				black.usage++
			case white.name:
				white.list = append(white.list, rgba)
				white.usage++
			case pink.name:
				pink.list = append(pink.list, rgba)
				pink.usage++
			case gray.name:
				gray.list = append(gray.list, rgba)
				gray.usage++
			case brown.name:
				brown.list = append(brown.list, rgba)
				brown.usage++
			}
		}
	}
	oldUsageList := []colorr{red, orange, yellow, green, blue, purple, black, white, pink, gray, brown}
	usageList := []colorr{}
	for _, i := range oldUsageList {
		if i.usage > 50 {
			usageList = append(usageList, i)
		}
	}
	if len(usageList) < 3 {
		return c.NewCaptcha()
	}
	sort.Sort(axisSorter(usageList))
	mostUsage := []colorr{}
	leastUsage := []colorr{}
	for i := 0; i < len(usageList)/2+1; i++ {
		if i < len(usageList)/2 {
			mostUsage = append(mostUsage, usageList[i])
		}
		leastUsage = append(leastUsage, usageList[len(usageList)-i-1])

	}

	coords := [][]int{}
	for _, i := range usageList {
		for _, j := range i.list {
			for _, k := range leastUsage {
				if i.name == k.name && !spedContains(coords, j) {
					coords = append(coords, j)
				}
			}
		}
	}
	//replaces the most used colors with least used colors
	for x := 1; x <= width; x++ {
		for y := 1; y <= height; y++ {
			colorr, rgba := findColor(img.At(x, y).RGBA())
			if rgba[3] > 200 && rand.Intn(7) == 0 {
				for _, yes := range mostUsage {
					if colorr == yes.name {
						colo := rand.Intn(len(coords))
						cimg.Set(x, y, color.RGBA{uint8(coords[colo][0]), uint8(coords[colo][1]), uint8(coords[colo][2]), uint8(coords[colo][3])})
					}
				}
			}
		}
	}
	aimg := image.NewRGBA(image.Rect(0, 0, width, height))
	//4-point warping
	var x1, y1, x2, y2, x3, y3, x4, y4, xa, ya, xb, yb, xc, yc, xd, yd int64
	//4 basis points
	x1, y1 = 1+int64(rand.Intn(width/9)), 1+int64(rand.Intn(height/9))
	x2, y2 = int64(width)-int64(rand.Intn(width/9)), 1+int64(rand.Intn(height/9))
	x3, y3 = int64(width)-int64(rand.Intn(width/9)), int64(height)-int64(rand.Intn(height/9))
	x4, y4 = 1+int64(rand.Intn(width/9)), int64(height)-int64(rand.Intn(height/9))
	//4 new points to tend to
	xa, ya = 1+int64(rand.Intn(width/3)), 1+int64(rand.Intn(height/3))
	xb, yb = int64(width)-int64(rand.Intn(width/3)), 1+int64(rand.Intn(height/3))
	xc, yc = int64(width)-int64(rand.Intn(width/3)), int64(height)-int64(rand.Intn(height/3))
	xd, yd = 1+int64(rand.Intn(width/3)), int64(height)-int64(rand.Intn(height/3))

	m := make([][]int64, 4)
	mm := make([][]int64, 4)
	m[0] = []int64{1, x1, y1, x1 * y1, xa}
	m[1] = []int64{1, x2, y2, x2 * y2, xb}
	m[2] = []int64{1, x3, y3, x3 * y3, xc}
	m[3] = []int64{1, x4, y4, x4 * y4, xd}
	mm[0] = []int64{1, x1, y1, x1 * y1, ya}
	mm[1] = []int64{1, x2, y2, x2 * y2, yb}
	mm[2] = []int64{1, x3, y3, x3 * y3, yc}
	mm[3] = []int64{1, x4, y4, x4 * y4, yd}
	m2 := make([][]rational.Rational, len(m))
	for i, iv := range m {
		mr := make([]rational.Rational, len(m[i]))
		for j, jv := range iv {
			mr[j] = rational.New(jv, 1)
		}
		m2[i] = mr
	}
	res, gausErr := solveGaussian(m2, false)
	if gausErr != nil {
		return nil, "", gausErr
	}
	for i, iv := range mm {
		mr := make([]rational.Rational, len(mm[i]))
		for j, jv := range iv {
			mr[j] = rational.New(jv, 1)
		}
		m2[i] = mr
	}
	res2, gausErr := solveGaussian(m2, false)
	if gausErr != nil {
		return nil, "", gausErr
	}
	xa2 := float64(res[0][0].Numerator) / float64(res[0][0].Denominator)
	xb2 := float64(res[1][0].Numerator) / float64(res[1][0].Denominator)
	xc2 := float64(res[2][0].Numerator) / float64(res[2][0].Denominator)
	xd2 := float64(res[3][0].Numerator) / float64(res[3][0].Denominator)

	ya2 := float64(res2[0][0].Numerator) / float64(res2[0][0].Denominator)
	yb2 := float64(res2[1][0].Numerator) / float64(res2[1][0].Denominator)
	yc2 := float64(res2[2][0].Numerator) / float64(res2[2][0].Denominator)
	yd2 := float64(res2[3][0].Numerator) / float64(res2[3][0].Denominator)

	for xx := 1; xx <= width; xx++ {
		for yy := 1; yy <= height; yy++ {
			_, rgba := findColor(cimg.At(xx, yy).RGBA())
			x := float64(xx)
			y := float64(yy)
			newx := xa2 + xb2*x + xc2*y + xd2*x*y
			newy := ya2 + yb2*x + yc2*y + yd2*x*y

			aimg.Set(abs(int(newx)), abs(int(newy)), color.RGBA{uint8(rgba[0]), uint8(rgba[1]), uint8(rgba[2]), uint8(rgba[3])})
		}
	}

	//adds lines using least used colors

	for lines := 0; lines < 7; lines++ {
		point1 := []float64{}
		if rand.Intn(2) == 0 {
			point1 = []float64{0, float64(rand.Intn(height) + 1)}
		} else {
			point1 = []float64{float64(rand.Intn(width) + 1), float64(height)}
		}
		point2 := []float64{}
		if rand.Intn(2) == 0 {
			point2 = []float64{float64(rand.Intn(width) + 1), 0}
		} else {
			point2 = []float64{float64(width), float64(rand.Intn(height) + 1)}
		}
		gc := draw2dimg.NewGraphicContext(aimg)
		colo := rand.Intn(len(coords))
		gc.SetStrokeColor(color.RGBA{uint8(coords[colo][0]), uint8(coords[colo][1]), uint8(coords[colo][2]), uint8(coords[colo][3])})
		gc.SetLineWidth((rand.Float64() * (.5)) + .1)
		gc.MoveTo(point1[0], point1[1])
		gc.LineTo(point2[0], point2[1])
		gc.Close()
		gc.FillStroke()
	}

	//adds circles using least used colors

	for circles := 0; circles < 3; circles++ {
		colo := rand.Intn(len(coords))
		drawCircle(aimg, rand.Intn(width), rand.Intn(height), rand.Intn(width/3)+height/3, color.RGBA{uint8(coords[colo][0]), uint8(coords[colo][1]), uint8(coords[colo][2]), uint8(coords[colo][3])})
	}
	img = imaging.AdjustSaturation(aimg, 50-float64(rand.Intn(100)))
	image := bytes.NewBuffer([]byte{})
	err = imaging.Encode(image, img, imaging.JPEG)
	if err != nil {
		return nil, "", err
	}
	return image, emoji, nil
}
func findColor(rr uint32, gg uint32, bb uint32, aa uint32) (string, []int) {
	r := int(rr / 257)
	g := int(gg / 257)
	b := int(bb / 257)
	a := int(aa / 257)
	min := math.Inf(1)
	var minColor string
	for _, i := range []colorr{red, orange, yellow, green, blue, purple, white, black, pink, gray, brown} {
		dist := dist(r, g, b, i)
		if i.name == black.name {
			dist = dist * 3
		}
		if dist < min {
			min = dist
			minColor = i.name
		}
	}
	return minColor, []int{r, g, b, a}
}

func dist(r int, g int, b int, colorr colorr) float64 {
	return float64((r-colorr.r)*(r-colorr.r) + (g-colorr.g)*(g-colorr.g) + (b-colorr.b)*(b-colorr.b))
}

func spedContains(coords [][]int, testcoords []int) bool {
	for _, i := range coords {
		if abs(i[0]-testcoords[0]) < 7 && abs(i[1]-testcoords[1]) < 7 && abs(i[2]-testcoords[2]) < 7 && abs(i[3]-testcoords[3]) < 7 {
			return true
		}
	}
	return false
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type axisSorter []colorr

func (a axisSorter) Len() int           { return len(a) }
func (a axisSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a axisSorter) Less(i, j int) bool { return a[i].usage > a[j].usage }

//thanks alex-ant
func solveGaussian(eqM [][]rational.Rational, printTriangularForm bool) (res [][]rational.Rational, err error) {
	if len(eqM) > len(eqM[0])-1 {
		err = errors.New("the number of equations can not be greater than the number of variables")
		return
	}

	dl, i, j := containsDuplicatesLines(eqM)
	if dl {
		err = fmt.Errorf("provided matrix contains duplicate lines (%d and %d)", i+1, j+1)
		return
	}

	for i := 0; i < len(eqM)-1; i++ {
		eqM = sortMatrix(eqM, i)

		var varC rational.Rational
		for k := i; k < len(eqM); k++ {
			if k == i {
				varC = eqM[k][i]
			} else {
				multipliedLine := make([]rational.Rational, len(eqM[i]))
				for z, zv := range eqM[i] {
					multipliedLine[z] = zv.Multiply(eqM[k][i].Divide(varC)).MultiplyByNum(-1)
				}
				newLine := make([]rational.Rational, len(eqM[k]))
				for z, zv := range eqM[k] {
					newLine[z] = zv.Add(multipliedLine[z])
				}
				eqM[k] = newLine
			}
		}
	}

	// Removing empty lines and inverting the matrix.
	var resultEqM [][]rational.Rational
	for i := len(eqM) - 1; i >= 0; i-- {
		if !rational.RationalsAreNull(eqM[i]) {
			resultEqM = append(resultEqM, eqM[i])
		}
	}

	getFirstNonZeroIndex := func(sl []rational.Rational) (index int) {
		for i, v := range sl {
			if v.GetNumerator() != 0 {
				index = i
				return
			}
		}
		return
	}

	// Back substitution.
	for z := 0; z < len(resultEqM)-1; z++ {
		var processIndex int
		var firstLine []rational.Rational
		for i := z; i < len(resultEqM); i++ {
			v := resultEqM[i]
			if i == z {
				processIndex = getFirstNonZeroIndex(v)
				firstLine = v
			} else {
				mult := v[processIndex].Divide(firstLine[processIndex]).MultiplyByNum(-1)
				for j, jv := range v {
					resultEqM[i][j] = firstLine[j].Multiply(mult).Add(jv)
				}
			}
		}
	}

	if printTriangularForm {
		for i := len(resultEqM) - 1; i >= 0; i-- {
			var str string
			for _, jv := range resultEqM[i] {
				str += strconv.FormatFloat(jv.Float64(), 'f', 2, 64) + ","
			}
			str = str[:len(str)-1]
			fmt.Println(str)
		}
	}

	// Calculating variables.
	res = make([][]rational.Rational, len(eqM[0])-1)
	if getFirstNonZeroIndex(resultEqM[0]) == len(resultEqM[0])-2 {
		// All the variables have been found.
		for i, iv := range resultEqM {
			index := len(res) - 1 - i
			res[index] = append(res[index], iv[len(iv)-1].Divide(iv[len(resultEqM)-1-i]))
		}
	} else {
		// Some variables remained unknown.
		var unknownStart, unknownEnd int
		for i, iv := range resultEqM {
			fnz := getFirstNonZeroIndex(iv)
			var firstRes []rational.Rational
			firstRes = append(firstRes, iv[len(iv)-1].Divide(iv[fnz]))
			if i == 0 {
				unknownStart = fnz + 1
				unknownEnd = len(iv) - 2
				for j := unknownEnd; j >= unknownStart; j-- {
					res[j] = []rational.Rational{rational.New(0, 0)}
					firstRes = append(firstRes, iv[j].Divide(iv[fnz]))
				}
			} else {
				for j := unknownEnd; j >= unknownStart; j-- {
					firstRes = append(firstRes, iv[j].Divide(iv[fnz]))
				}
			}
			res[fnz] = firstRes
		}
	}

	return
}

func sortMatrix(m [][]rational.Rational, initRow int) (m2 [][]rational.Rational) {
	indexed := make(map[int]bool)

	for i := 0; i < initRow; i++ {
		m2 = append(m2, m[i])
		indexed[i] = true
	}

	greaterThanMax := func(rr1, rr2 []rational.Rational) (greater bool) {
		for i := 0; i < len(rr1); i++ {
			if rr1[i].GetModule().GreaterThan(rr2[i].GetModule()) {
				greater = true
				return
			} else if rr1[i].GetModule().LessThan(rr2[i].GetModule()) {
				return
			}
		}
		return
	}

	type maxStruct struct {
		index   int
		element []rational.Rational
	}

	for i := initRow; i < len(m); i++ {
		max := maxStruct{-1, make([]rational.Rational, len(m[i]))}
		var firstNotIndexed int
		for k, kv := range m {
			if !indexed[k] {
				firstNotIndexed = k
				if greaterThanMax(kv, max.element) {
					max.index = k
					max.element = kv
				}
			}
		}
		if max.index != -1 {
			m2 = append(m2, max.element)
			indexed[max.index] = true
		} else {
			m2 = append(m2, m[firstNotIndexed])
			indexed[firstNotIndexed] = true
		}
	}

	return
}

func containsDuplicatesLines(eqM [][]rational.Rational) (contains bool, l1, l2 int) {
	for i := 0; i < len(eqM); i++ {
		for j := i + 1; j < len(eqM); j++ {
			var equalElements int
			for k := 0; k < len(eqM[i]); k++ {
				if eqM[i][k] == eqM[j][k] {
					equalElements++
				} else {
					break
				}
			}
			if equalElements == len(eqM[i]) {
				contains = true
				l1 = i
				l2 = j
				return
			}
		}
	}
	return
}
func drawCircle(aimg draw.Image, x0, y0, r int, c color.RGBA) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x > y {
		aimg.Set(x0+x, y0+y, c)
		aimg.Set(x0+y, y0+x, c)
		aimg.Set(x0-y, y0+x, c)
		aimg.Set(x0-x, y0+y, c)
		aimg.Set(x0-x, y0-y, c)
		aimg.Set(x0-y, y0-x, c)
		aimg.Set(x0+y, y0-x, c)
		aimg.Set(x0+x, y0-y, c)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
}
