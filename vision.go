package main

import "math"

const ViewRadius = 5

type VisionMap struct {
	Visible  [][]bool
	Explored [][]bool
	width    int
	height   int
}

func NewVisionMap(w, h int) *VisionMap {
	v := &VisionMap{
		Visible:  make([][]bool, h),
		Explored: make([][]bool, h),
		width:    w,
		height:   h,
	}
	for y := 0; y < h; y++ {
		v.Visible[y] = make([]bool, w)
		v.Explored[y] = make([]bool, w)
	}
	return v
}

func (v *VisionMap) ClearVisible() {
	for y := 0; y < v.height; y++ {
		for x := 0; x < v.width; x++ {
			v.Visible[y][x] = false
		}
	}
}

func (v *VisionMap) IsInBounds(x, y int) bool {
	return x >= 0 && x < v.width && y >= 0 && y < v.height
}

func (v *VisionMap) ComputeFOV(gameMap [][]Tile, px, py int, radius int) {
	v.ClearVisible()

	v.setVisibleInternal(px, py)

	numRays := 360
	for i := 0; i < numRays; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numRays)
		dx := math.Cos(angle)
		dy := math.Sin(angle)
		v.castRay(gameMap, px, py, dx, dy, radius)
	}
}

func (v *VisionMap) castRay(gameMap [][]Tile, px, py int, dx, dy float64, radius int) {
	x := float64(px) + 0.5
	y := float64(py) + 0.5
	stepSize := 0.1
	maxSteps := int(float64(radius) / stepSize)

	for i := 0; i < maxSteps; i++ {
		x += dx * stepSize
		y += dy * stepSize

		ix := int(x)
		iy := int(y)

		if !v.IsInBounds(ix, iy) {
			return
		}

		distX := float64(ix - px)
		distY := float64(iy - py)
		if distX*distX+distY*distY > float64(radius*radius) {
			return
		}

		v.setVisibleInternal(ix, iy)

		if !gameMap[iy][ix].Walkable {
			return
		}
	}
}

func (v *VisionMap) setVisibleInternal(x, y int) {
	if v.IsInBounds(x, y) {
		v.Visible[y][x] = true
		v.Explored[y][x] = true
	}
}

func (v *VisionMap) IsVisible(x, y int) bool {
	if !v.IsInBounds(x, y) {
		return false
	}
	return v.Visible[y][x]
}

func (v *VisionMap) IsExplored(x, y int) bool {
	if !v.IsInBounds(x, y) {
		return false
	}
	return v.Explored[y][x]
}
