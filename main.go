package main

import (
	"image"
	"image/draw"
	"path/filepath"
	"runtime"
	"time"

	g143 "github.com/bankole7782/graphics143"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	fps              = 10
	fontSize         = 20
	toolBoxW         = 150
	toolBoxH         = 40
	indicatorCircleR = 8

	PencilWidget        = 101
	CanvasWidget        = 102
	SymmLineWidget      = 103
	LeftSymmWidget      = 104
	RefLineWidget       = 105
	ClearRefLinesWidget = 106
	SaveWidget          = 107
)

// var objCoords map[g143.RectSpecs]any
var objCoords map[int]g143.RectSpecs

var currentWindowFrame image.Image

type CircleSpec struct {
	X int
	Y int
}

var drawnIndicators []CircleSpec
var activeTool int
var lastX, lastY float64  // used in drawing
var lastSymmLineX float64 // used in drawing
var refLinesDrawnPosY []float64

func main() {
	runtime.LockOSThread()

	objCoords = make(map[int]g143.RectSpecs)
	drawnIndicators = make([]CircleSpec, 0)
	refLinesDrawnPosY = make([]float64, 0)

	window := g143.NewWindow(1450, 700, "images376: a 3d reference image creator. Majoring on faces", false)
	allDraws(window)

	// respond to the mouse
	window.SetMouseButtonCallback(mouseBtnCallback)
	// respond to mouse movement
	window.SetCursorPosCallback(cursorPosCallback)

	for !window.ShouldClose() {
		t := time.Now()
		glfw.PollEvents()

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}

}

func allDraws(window *glfw.Window) {
	wWidth, wHeight := window.GetSize()

	// frame buffer
	ggCtx := gg.NewContext(wWidth, wHeight)

	// background rectangle
	ggCtx.DrawRectangle(0, 0, float64(wWidth), float64(wHeight))
	ggCtx.SetHexColor("#ddd")
	ggCtx.Fill()

	// intro text
	FontPath := getDefaultFontPath()
	// draw the tools
	err := ggCtx.LoadFontFace(FontPath, fontSize)
	if err != nil {
		panic(err)
	}

	// draw tools box
	ggCtx.SetHexColor("#DAC166")
	ggCtx.DrawRoundedRectangle(10, 10, toolBoxW+20, 320, 10)
	ggCtx.Fill()

	// pencil tool
	ggCtx.SetHexColor("#dddddd")
	ggCtx.DrawRectangle(20, 20, toolBoxW, toolBoxH)
	ggCtx.Fill()

	pencilRS := g143.RectSpecs{Width: toolBoxW, Height: toolBoxH, OriginX: 20, OriginY: 20}
	objCoords[PencilWidget] = pencilRS

	ggCtx.SetHexColor("#444444")
	ggCtx.DrawString("Pencil", 30, 30+fontSize)

	// symm line tool
	ggCtx.SetHexColor("#dddddd")
	ggCtx.DrawRectangle(20, 70, toolBoxW, toolBoxH)
	ggCtx.Fill()

	slRS := g143.RectSpecs{Width: toolBoxW, Height: toolBoxH, OriginX: 20, OriginY: 70}
	objCoords[SymmLineWidget] = slRS

	ggCtx.SetHexColor("#444444")
	ggCtx.DrawString("Symm Line", 30, 80+fontSize)

	// Left symm tool
	ggCtx.SetHexColor("#dddddd")
	lswY := slRS.OriginY + slRS.Height + 10
	ggCtx.DrawRectangle(20, float64(lswY), toolBoxW, toolBoxH)
	ggCtx.Fill()
	lsRS := g143.RectSpecs{Width: toolBoxW, Height: toolBoxH, OriginX: 20, OriginY: lswY}
	objCoords[LeftSymmWidget] = lsRS

	ggCtx.SetHexColor("#444444")
	ggCtx.DrawString("Left Symm", 30, float64(lsRS.OriginY)+fontSize+10)

	// Refline tool
	ggCtx.SetHexColor("#ddd")
	rlwY := lsRS.OriginY + lsRS.Height + 10
	ggCtx.DrawRectangle(20, float64(rlwY), toolBoxW, toolBoxH)
	ggCtx.Fill()
	rlRS := g143.RectSpecs{Width: toolBoxW, Height: toolBoxH, OriginX: 20, OriginY: rlwY}
	objCoords[RefLineWidget] = rlRS

	ggCtx.SetHexColor("#444")
	ggCtx.DrawString("Ref Line", 30, float64(rlRS.OriginY)+fontSize+10)

	// save tool
	ggCtx.SetHexColor("#ddd")
	swY := rlRS.OriginY + rlRS.Height + 10
	ggCtx.DrawRectangle(20, float64(swY), toolBoxW, toolBoxH)
	ggCtx.Fill()
	swRS := g143.RectSpecs{Width: toolBoxW, Height: toolBoxH, OriginX: 20, OriginY: swY}
	objCoords[SaveWidget] = swRS

	ggCtx.SetHexColor("#444")
	ggCtx.DrawString("Save Ref", 30, float64(swRS.OriginY)+fontSize+10)

	// Canvas
	ggCtx.SetHexColor("#ffffff")
	ggCtx.DrawRectangle(200, 10, 1200, 600)
	ggCtx.Fill()

	canvasRS := g143.RectSpecs{Width: 1200, Height: 600, OriginX: 200, OriginY: 10}
	objCoords[CanvasWidget] = canvasRS

	// draw divider
	ggCtx.SetHexColor("#444")
	ggCtx.SetLineWidth(2)
	ggCtx.MoveTo(float64(canvasRS.OriginX)+float64(canvasRS.Width)/2, float64(canvasRS.OriginY))
	ggCtx.LineTo(float64(canvasRS.OriginX)+float64(canvasRS.Width)/2, float64(canvasRS.OriginY)+float64(canvasRS.Height))
	ggCtx.Stroke()

	// send the frame to glfw window
	windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
	g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
	window.SwapBuffers()

	// save the frame
	currentWindowFrame = ggCtx.Image()
}

func mouseBtnCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Release {
		return
	}

	xPos, yPos := window.GetCursorPos()
	xPosInt := int(xPos)
	yPosInt := int(yPos)

	wWidth, wHeight := window.GetSize()

	var widgetRS g143.RectSpecs
	var widgetCode int

	for code, RS := range objCoords {
		if g143.InRectSpecs(RS, xPosInt, yPosInt) {
			widgetRS = RS
			widgetCode = code
			break
		}
	}

	if widgetCode == 0 {
		return
	}

	switch widgetCode {
	case PencilWidget, SymmLineWidget, RefLineWidget:

		ggCtx := gg.NewContextForImage(currentWindowFrame)

		activeTool = widgetCode

		// clear indicators
		for _, cs := range drawnIndicators {
			ggCtx.SetHexColor("#dddddd")
			ggCtx.DrawCircle(float64(cs.X), float64(cs.Y), indicatorCircleR+2)
			ggCtx.Fill()
		}
		// draw an indicator on the active tool
		ggCtx.SetHexColor("#DAC166")
		ggCtx.DrawCircle(float64(widgetRS.OriginX+widgetRS.Width-20), float64(widgetRS.OriginY+20), 10)
		ggCtx.Fill()
		drawnIndicators = append(drawnIndicators, CircleSpec{X: widgetRS.OriginX + widgetRS.Width - 20, Y: widgetRS.OriginY + 20})

		// send the frame to glfw window
		windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
		window.SwapBuffers()

		// save the frame
		currentWindowFrame = ggCtx.Image()

	case SaveWidget:
		ggCtx := gg.NewContextForImage(currentWindowFrame)
		activeTool = 0

		// clear indicators
		for _, cs := range drawnIndicators {
			ggCtx.SetHexColor("#dddddd")
			ggCtx.DrawCircle(float64(cs.X), float64(cs.Y), indicatorCircleR+2)
			ggCtx.Fill()
		}

		canvasRS := objCoords[CanvasWidget]

		newImageRect := image.Rect(0, 0, canvasRS.Width, canvasRS.Height)
		outImg := image.NewRGBA(newImageRect)
		draw.Draw(outImg, newImageRect, currentWindowFrame, image.Pt(canvasRS.OriginX, canvasRS.OriginY), draw.Src)

		rootPath, _ := GetRootPath()
		outPath := filepath.Join(rootPath, time.Now().Format("20060102T150405MST")+".png")
		imaging.Save(outImg, outPath)

	case CanvasWidget:

		ggCtx := gg.NewContextForImage(currentWindowFrame)
		ctrlState := window.GetKey(glfw.KeyLeftControl)
		canvasRS := objCoords[CanvasWidget]

		// SymLine Widget
		if activeTool == SymmLineWidget && ctrlState == glfw.Release {
			// clear last symmline
			if int(lastSymmLineX) != 0 {
				ggCtx.SetHexColor("#fff")
				ggCtx.SetLineWidth(2)
				ggCtx.MoveTo(lastSymmLineX, float64(canvasRS.OriginY))
				ggCtx.LineTo(lastSymmLineX, float64(canvasRS.Height+canvasRS.OriginY))
				ggCtx.Stroke()
			}

			// symline widget should only work in the left axis
			if xPos > (float64(canvasRS.OriginX) + float64(canvasRS.Width/2)) {
				return
			}
			ggCtx.SetHexColor("#999")
			ggCtx.SetLineWidth(1)
			ggCtx.MoveTo(xPos, float64(canvasRS.OriginY))
			ggCtx.LineTo(xPos, float64(canvasRS.Height+canvasRS.OriginY))
			ggCtx.Stroke()

			lastSymmLineX = xPos

		} else if activeTool == SymmLineWidget && ctrlState == glfw.Press {
			// clear last symmline
			if int(lastSymmLineX) != 0 {
				ggCtx.SetHexColor("#fff")
				ggCtx.SetLineWidth(2)
				ggCtx.MoveTo(lastSymmLineX, float64(canvasRS.OriginY))
				ggCtx.LineTo(lastSymmLineX, float64(canvasRS.Height+canvasRS.OriginY))
				ggCtx.Stroke()

			}

			lastSymmLineX = 0
		}

		// Reference Line Widget
		if activeTool == RefLineWidget && ctrlState == glfw.Release {

			ggCtx.SetHexColor(GetRandomColorInHex())
			ggCtx.SetLineWidth(1)
			ggCtx.MoveTo(float64(canvasRS.OriginX), yPos)
			ggCtx.LineTo(float64(canvasRS.OriginX+canvasRS.Width), yPos)
			ggCtx.Stroke()

			refLinesDrawnPosY = append(refLinesDrawnPosY, yPos)

		} else if activeTool == RefLineWidget && ctrlState == glfw.Press {

			// clear old ref lines
			for _, oldRefLinePos := range refLinesDrawnPosY {
				ggCtx.SetHexColor("#fff")
				ggCtx.SetLineWidth(2)
				ggCtx.MoveTo(float64(canvasRS.OriginX), oldRefLinePos)
				ggCtx.LineTo(float64(canvasRS.OriginX+canvasRS.Width), oldRefLinePos)
				ggCtx.Stroke()
			}

		}

		// send the frame to glfw window
		windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
		window.SwapBuffers()

		// save the frame
		currentWindowFrame = ggCtx.Image()

	case LeftSymmWidget:
		canvasRS := objCoords[CanvasWidget]
		ggCtx := gg.NewContextForImage(currentWindowFrame)

		if lastSymmLineX == 0 {
			return
		}
		// clear last symmLine
		ggCtx.SetHexColor("#fff")
		ggCtx.SetLineWidth(2)
		ggCtx.MoveTo(lastSymmLineX, float64(canvasRS.OriginY))
		ggCtx.LineTo(lastSymmLineX, float64(canvasRS.Height+canvasRS.OriginY))
		ggCtx.Stroke()

		// begin left symmetrize
		leftHalfRect := image.Rect(0, 0, int(lastSymmLineX)-canvasRS.OriginX, canvasRS.Height)
		leftHalfImg := image.NewRGBA(leftHalfRect)
		draw.Draw(leftHalfImg, leftHalfRect, currentWindowFrame, image.Pt(canvasRS.OriginX, canvasRS.OriginY), draw.Src)

		tmpLeftHalfImg := imaging.FlipH(leftHalfImg)

		tmpFullRect := image.Rect(0, 0, 2*leftHalfRect.Dx(), canvasRS.Height)
		tmpFullImg := image.NewRGBA(tmpFullRect)

		rightHalfRect := image.Rect(leftHalfRect.Dx(), 0, 2*leftHalfRect.Dx(), canvasRS.Height)
		draw.Draw(tmpFullImg, leftHalfRect, leftHalfImg, image.Point{}, draw.Src)
		draw.Draw(tmpFullImg, rightHalfRect, tmpLeftHalfImg, image.Point{}, draw.Src)

		tmpFullImg2 := imaging.Crop(tmpFullImg.SubImage(tmpFullRect), image.Rect(0, 0, canvasRS.Width/2-2, canvasRS.Height))
		ggCtx.DrawImage(tmpFullImg2, canvasRS.OriginX, canvasRS.OriginY)

		// clear active tool selection
		activeTool = 0
		// clear indicators
		for _, cs := range drawnIndicators {
			ggCtx.SetHexColor("#dddddd")
			ggCtx.DrawCircle(float64(cs.X), float64(cs.Y), indicatorCircleR+2)
			ggCtx.Fill()
		}

		// send the frame to glfw window
		windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
		window.SwapBuffers()

		// save the frame
		currentWindowFrame = ggCtx.Image()

	default:

	}
}

func cursorPosCallback(window *glfw.Window, xpos float64, ypos float64) {
	wWidth, wHeight := window.GetSize()

	ggCtx := gg.NewContextForImage(currentWindowFrame)
	canvasRS := objCoords[CanvasWidget]

	currentMouseAction := window.GetMouseButton(glfw.MouseButtonLeft)

	if currentMouseAction == glfw.Release {
		lastX, lastY = 0.0, 0.0
	}

	ctrlState := window.GetKey(glfw.KeyLeftControl)

	if g143.InRectSpecs(canvasRS, int(xpos), int(ypos)) && currentMouseAction == glfw.Press {

		// Pencil Widget
		if activeTool == PencilWidget && ctrlState == glfw.Release {
			// draw circles
			ggCtx.SetHexColor("#222222")

			if int(lastX) != 0 {
				ggCtx.SetLineWidth(4)
				ggCtx.MoveTo(lastX, lastY)
				ggCtx.LineTo(xpos, ypos)
				ggCtx.Stroke()
			} else {
				ggCtx.DrawCircle(xpos, ypos, 3)
				ggCtx.Fill()
			}

			lastX, lastY = xpos, ypos
		} else if activeTool == PencilWidget && ctrlState == glfw.Press {
			ggCtx.SetHexColor("#ffffff")
			ggCtx.DrawCircle(xpos, ypos, 10)
			ggCtx.Fill()
		}

	}

	// send the frame to glfw window
	windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
	g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
	window.SwapBuffers()

	// save the frame
	currentWindowFrame = ggCtx.Image()
}
