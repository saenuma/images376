package main

import (
	"image"
	"image/draw"
	"os/exec"
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
	canvasWidth      = 1200
	canvasHeight     = 600

	PencilWidget        = 101
	CanvasWidget        = 102
	SymmLineWidget      = 103
	LeftSymmWidget      = 104
	RefLineWidget       = 105
	ClearRefLinesWidget = 106
	SaveWidget          = 107
	OpenWDWidget        = 108
)

// var objCoords map[g143.RectSpecs]any
var objCoords map[int]g143.RectSpecs

type CircleSpec struct {
	X int
	Y int
}

var drawnIndicators []CircleSpec
var activeTool int
var lastX, lastY float64  // used in drawing
var lastSymmLineX float64 // used in drawing

// images
var currentWindowFrame image.Image
var pencilLayerImg image.Image
var linesLayerImg image.Image

func main() {
	runtime.LockOSThread()

	// white image in pencilLayerImg
	ggCtx := gg.NewContext(canvasWidth, canvasHeight)
	ggCtx.DrawRectangle(0, 0, float64(canvasWidth), float64(canvasHeight))
	ggCtx.SetHexColor("#fff")
	ggCtx.Fill()
	pencilLayerImg = ggCtx.Image()

	linesLayerImg = image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	objCoords = make(map[int]g143.RectSpecs)
	drawnIndicators = make([]CircleSpec, 0)

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

	// Open Outputs
	ggCtx.SetHexColor("#ddd")
	ooY := swRS.OriginY + swRS.Height + 10
	ggCtx.DrawRectangle(20, float64(ooY), toolBoxW, toolBoxH)
	ggCtx.Fill()
	ooRS := g143.RectSpecs{Width: toolBoxW, Height: toolBoxH, OriginX: 20, OriginY: ooY}
	objCoords[OpenWDWidget] = ooRS

	ggCtx.SetHexColor("#444")
	ggCtx.DrawString("Open Folder", 30, float64(ooRS.OriginY)+fontSize+10)

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

	// write indicators
	ggCtx.SetHexColor("#444")
	indicatorsY := canvasRS.OriginY + canvasRS.Height + 20
	ggCtx.DrawString("Front View", toolBoxW+300, float64(indicatorsY)+fontSize)
	ggCtx.DrawString("Side View", toolBoxW+300+canvasWidth/2, float64(indicatorsY)+fontSize)

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

	rootPath, _ := GetRootPath()
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

		rootPath, _ := GetRootPath()
		outPath := filepath.Join(rootPath, time.Now().Format("20060102T150405MST")+".png")
		imaging.Save(pencilLayerImg, outPath)

		// send the frame to glfw window
		windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
		window.SwapBuffers()

		// save the frame
		currentWindowFrame = ggCtx.Image()

	case CanvasWidget:

		ggCtx := gg.NewContextForImage(currentWindowFrame)
		ctrlState := window.GetKey(glfw.KeyLeftControl)
		canvasRS := objCoords[CanvasWidget]

		linesLayerggCtx := gg.NewContextForImage(linesLayerImg)
		translastedMouseX, translatedMouseY := xPos-float64(canvasRS.OriginX), yPos-float64(canvasRS.OriginY)

		// SymLine Widget
		if activeTool == SymmLineWidget && ctrlState == glfw.Release {
			// clear last symmline
			if int(lastSymmLineX) != 0 {
				linesLayerImg = image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
				linesLayerggCtx = gg.NewContextForImage(linesLayerImg)
			}

			// symline widget should only work in the left axis
			if xPos > (float64(canvasRS.OriginX) + float64(canvasRS.Width/2)) {
				return
			}

			linesLayerggCtx.SetHexColor("#999")
			linesLayerggCtx.SetLineWidth(1)
			linesLayerggCtx.MoveTo(translastedMouseX, 0)
			linesLayerggCtx.LineTo(translastedMouseX, float64(canvasRS.Height))
			linesLayerggCtx.Stroke()

			lastSymmLineX = translastedMouseX

		} else if activeTool == SymmLineWidget && ctrlState == glfw.Press {
			linesLayerImg = image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
			linesLayerggCtx = gg.NewContextForImage(linesLayerImg)

			lastSymmLineX = 0
		}

		// Reference Line Widget
		if activeTool == RefLineWidget && ctrlState == glfw.Release {

			linesLayerggCtx.SetHexColor(GetRandomColorInHex())
			linesLayerggCtx.SetLineWidth(1)
			linesLayerggCtx.MoveTo(0, translatedMouseY)
			linesLayerggCtx.LineTo(float64(canvasRS.Width), translatedMouseY)
			linesLayerggCtx.Stroke()

			linesLayerImg = linesLayerggCtx.Image()

		} else if activeTool == RefLineWidget && ctrlState == glfw.Press {

			// clear old ref lines
			linesLayerImg = image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
			linesLayerggCtx = gg.NewContextForImage(linesLayerImg)
		}

		ggCtx.DrawImage(pencilLayerImg, canvasRS.OriginX, canvasRS.OriginY)
		ggCtx.DrawImage(linesLayerggCtx.Image(), canvasRS.OriginX, canvasRS.OriginY)

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

	case LeftSymmWidget:
		canvasRS := objCoords[CanvasWidget]
		ggCtx := gg.NewContextForImage(currentWindowFrame)

		if lastSymmLineX == 0 {
			return
		}
		// clear last symmLine
		linesLayerImg = image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

		// begin left symmetrize
		leftHalfRect := image.Rect(0, 0, int(lastSymmLineX), canvasRS.Height)
		leftHalfImg := image.NewRGBA(leftHalfRect)
		draw.Draw(leftHalfImg, leftHalfRect, pencilLayerImg, image.Point{}, draw.Src)

		tmpLeftHalfImg := imaging.FlipH(leftHalfImg)

		tmpFullRect := image.Rect(0, 0, 2*leftHalfRect.Dx(), canvasRS.Height)
		tmpFullImg := image.NewRGBA(tmpFullRect)

		rightHalfRect := image.Rect(leftHalfRect.Dx(), 0, 2*leftHalfRect.Dx(), canvasRS.Height)
		draw.Draw(tmpFullImg, leftHalfRect, leftHalfImg, image.Point{}, draw.Src)
		draw.Draw(tmpFullImg, rightHalfRect, tmpLeftHalfImg, image.Point{}, draw.Src)

		tmpFullImg2 := imaging.Crop(tmpFullImg.SubImage(tmpFullRect), image.Rect(0, 0, canvasRS.Width/2-2, canvasRS.Height))

		pencilLayerggCtx := gg.NewContextForImage(pencilLayerImg)
		pencilLayerggCtx.DrawImage(tmpFullImg2, 0, 0)
		pencilLayerImg = pencilLayerggCtx.Image()

		ggCtx.DrawImage(pencilLayerggCtx.Image(), canvasRS.OriginX, canvasRS.OriginY)
		// draw divider
		ggCtx.SetHexColor("#444")
		ggCtx.SetLineWidth(2)
		ggCtx.MoveTo(float64(canvasRS.OriginX)+float64(canvasRS.Width)/2, float64(canvasRS.OriginY))
		ggCtx.LineTo(float64(canvasRS.OriginX)+float64(canvasRS.Width)/2, float64(canvasRS.OriginY)+float64(canvasRS.Height))
		ggCtx.Stroke()

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

	case OpenWDWidget:
		if runtime.GOOS == "windows" {
			exec.Command("cmd", "/C", "start", rootPath).Run()
		} else if runtime.GOOS == "linux" {
			exec.Command("xdg-open", rootPath).Run()
		}
	default:

	}
}

var count = 0

func cursorPosCallback(window *glfw.Window, xpos float64, ypos float64) {
	if runtime.GOOS == "linux" {
		// linux fires too many events
		count += 1
		if count != 10 {
			return
		} else {
			count = 0
		}
	}
	wWidth, wHeight := window.GetSize()

	ggCtx := gg.NewContextForImage(currentWindowFrame)
	canvasRS := objCoords[CanvasWidget]

	pencilLayerggCtx := gg.NewContextForImage(pencilLayerImg)

	currentMouseAction := window.GetMouseButton(glfw.MouseButtonLeft)

	if currentMouseAction == glfw.Release {
		lastX, lastY = 0.0, 0.0
	}

	ctrlState := window.GetKey(glfw.KeyLeftControl)

	if g143.InRectSpecs(canvasRS, int(xpos), int(ypos)) && currentMouseAction == glfw.Press {

		// Pencil Widget
		translastedMouseX, translatedMouseY := xpos-float64(canvasRS.OriginX), ypos-float64(canvasRS.OriginY)
		if activeTool == PencilWidget && ctrlState == glfw.Release && int(lastX) != 0 {
			// draw circles
			pencilLayerggCtx.SetHexColor("#222222")

			pencilLayerggCtx.SetLineWidth(4)
			pencilLayerggCtx.MoveTo(lastX, lastY)
			pencilLayerggCtx.LineTo(translastedMouseX, translatedMouseY)
			pencilLayerggCtx.Stroke()

		} else if activeTool == PencilWidget && ctrlState == glfw.Press && int(lastX) != 0 {
			pencilLayerggCtx.SetHexColor("#ffffff")
			pencilLayerggCtx.SetLineWidth(20)
			pencilLayerggCtx.MoveTo(lastX, lastY)
			pencilLayerggCtx.LineTo(translastedMouseX, translatedMouseY)
			pencilLayerggCtx.Stroke()
		}

		lastX, lastY = translastedMouseX, translatedMouseY

		pencilLayerImg = pencilLayerggCtx.Image()
		ggCtx.DrawImage(pencilLayerggCtx.Image(), canvasRS.OriginX, canvasRS.OriginY)
		ggCtx.DrawImage(linesLayerImg, canvasRS.OriginX, canvasRS.OriginY)

		// draw divider
		ggCtx.SetHexColor("#444")
		ggCtx.SetLineWidth(2)
		ggCtx.MoveTo(float64(canvasRS.OriginX)+float64(canvasRS.Width)/2, float64(canvasRS.OriginY))
		ggCtx.LineTo(float64(canvasRS.OriginX)+float64(canvasRS.Width)/2, float64(canvasRS.OriginY)+float64(canvasRS.Height))
		ggCtx.Stroke()

	}

	// send the frame to glfw window
	windowRS := g143.RectSpecs{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
	g143.DrawImage(wWidth, wHeight, ggCtx.Image(), windowRS)
	window.SwapBuffers()

	// save the frame
	currentWindowFrame = ggCtx.Image()
}
