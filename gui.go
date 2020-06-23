package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/ptool"
	browser "github.com/malashin/dialog"
)

func setupUI(app fyne.App) {
	win = app.NewWindow("TreeSnap")
	win.SetFixedSize(true)
	pBar = widget.NewProgressBar()
	pBar.Hide()

	win.SetContent(
		widget.NewVBox(
			widget.NewButton("Select focus file(s)", func() { selectFocusFiles() }),
			widget.NewButton("Select HOI4 folder", func() { selectGameFolder() }),
			widget.NewButton("Add dependency mod folder(s)", func() { selectModFolder() }),
			widget.NewButton("Select localisation language", func() { selectLocLanguage(app) }),
			widget.NewButton("Generate image", func() { start() }),
			widget.NewCheck("Disable line rendering", func(on bool) { lineRenderingToggle(on) }),
			pBar,
			widget.NewButton("Quit", func() {
				app.Quit()
			}),
		),
	)

	win.ShowAndRun()
}

func selectFocusFiles() {
	filename, err := browser.File().Title("National Focus File").Filter("Text file", "txt").LoadFiles()
	if err != nil {
		if err.Error() == "Cancelled" {
			return
		}
		ansi.Println("\x1b[31;1m" + err.Error() + "\x1b[0m")
		showError(err)
		return
	}
	focusTreePaths = filename
	ansi.Println("Focus files selected:", filename)
}

func selectGameFolder() {
	directory, err := browser.Directory().Title("HOI4 Folder").Browse()
	if err != nil {
		if err.Error() == "Cancelled" {
			return
		}
		ansi.Println("\x1b[31;1m" + err.Error() + "\x1b[0m")
		showError(err)
		return
	}
	gamePath = directory
	ansi.Println("Game folder selected:", directory)
	err = encodeCacheFile(gamePath, filepath.Join(binPath, "hoi4treesnapGamePath.txt"))
	if err != nil && err.Error() != "Cancelled" {
		ansi.Println("\x1b[31;1m" + err.Error() + "\x1b[0m")
		showError(err)
		return
	}
}

func selectModFolder() {
	directory, err := browser.Directory().Title("Mod Folder").Browse()
	if err != nil {
		if err.Error() == "Cancelled" {
			return
		}
		ansi.Println("\x1b[31;1m" + err.Error() + "\x1b[0m")
		showError(err)
		return
	}
	modPaths = append(modPaths, directory)
	ansi.Println("Mod folder added:", directory)
}

func selectLocLanguage(app fyne.App) {
	w := app.NewWindow("Select localisation language")
	w.SetFixedSize(true)

	w.SetContent(
		widget.NewVBox(
			widget.NewRadio([]string{"English", "Brazilian Portuguese", "German", "French", "Spanish", "Polish", "Russian"}, func(s string) { handleLocLanguageChange(s, w) }),
		),
	)

	w.Show()
}

func handleLocLanguageChange(s string, w fyne.Window) {
	switch s {
	case "English":
		language = "l_english"
	case "Brazilian Portuguese":
		language = "l_braz_por"
	case "German":
		language = "l_german"
	case "French":
		language = "l_french"
	case "Spanish":
		language = "l_spanish"
	case "Polish":
		language = "l_polish"
	case "Russian":
		language = "l_russian"
	}
	ansi.Println("Language selected:", s)
	w.Close()
}

func lineRenderingToggle(on bool) {
	if on {
		isLineRenderingOff = true
	} else {
		isLineRenderingOff = false
	}
}

func start() {
	if running {
		return
	}
	running = true

	var err error
	locMap[language] = make(map[string]Localisation)
	gfxList = append(gfxList, "GFX_focus_can_start")

	switch {
	case len(focusTreePaths) == 0:
		showError(errors.New("Focus file not selected"))
		return
	case gamePath == "":
		p := filepath.Join(binPath, "hoi4treesnapGamePath.txt")
		if _, err = os.Stat(p); err == nil {
			err = decodeCacheFile(&gamePath, p)
			if err != nil {
				showError(err)
				return
			}
		} else {
			showError(errors.New("Game path not selected"))
			return
		}
	}

	// Track start time for benchmarking.
	startTime := time.Now()

	// Build parsers.
	pdx, err = ptool.NewBuilder().FromString(pdxRule).Entries("entry").Build()
	if err != nil {
		showError(err)
		return
	}
	yml, err = ptool.NewBuilder().FromString(ymlRule).Entries("entry").Build()
	if err != nil {
		showError(err)
		return
	}

	for _, focusTreePath := range focusTreePaths {
		// Show progress bar.
		pBar.Show()

		focusTreeName := filepath.Base(focusTreePath)
		focusTreeName = focusTreeName[0 : len(focusTreeName)-len(filepath.Ext(focusTreeName))]

		modPath := filepath.Clean(strings.TrimSuffix(filepath.Dir(focusTreePath), filepath.Join("common", "national_focus")))
		// Add gamePath to the front of modsPath slice.
		if !containsString(modPaths, gamePath) {
			modPaths = append([]string{gamePath}, modPaths...)
		}
		// If modsPaths slice does not contain the mod path the focus tree is in add it to the end of the slice.
		if !containsString(modPaths, modPath) {
			modPaths = append(modPaths, modPath)
		}

		ansi.Println("\x1b[33;1m" + "Parsing files:" + "\x1b[0m")
		// Focus tree parsing.
		err = parseFocus(focusTreePath)
		if err != nil {
			showError(err)
			return
		}
		pBar.SetValue(0.05)

		// Parse focus tree gui.
		// Find the last nationalfocusview.gui in the modPaths slice.
		guiPath := gamePath
		if len(modPaths) > 1 {
			for _, p := range modPaths[1:] {
				if _, err = os.Stat(filepath.Join(p, "interface", "nationalfocusview.gui")); err == nil {
					guiPath = p
				}
			}
		}
		err = parseGUI(guiPath)
		if err != nil {
			showError(err)
			return
		}
		pBar.SetValue(0.1)

		// GFX parsing.
		for _, p := range modPaths {
			err = parseGFX(p, len(modPaths))
			if err != nil {
				showError(err)
				return
			}
		}

		// Parse localisation files.
		for _, p := range modPaths {
			err = parseLoc(p, len(modPaths))
			if err != nil {
				showError(err)
				return
			}
		}

		ansi.Println("\x1b[33;1m" + "Generating images:" + "\x1b[0m")
		var i float64 = 8
		// Replace hoi4 textures if mods has the same ones.
		useModsTexturesIfPresent()
		pBar.SetValue(pBar.Value + 0.1/i)

		// Calculate coordinates of focuses with relative positions.
		fillAbsoluteFocusPositions(true)
		pBar.SetValue(pBar.Value + 0.1/i)

		// Fill in focus structs with children data.
		fillFocusChildAndParentData()
		pBar.SetValue(pBar.Value + 0.1/i)

		// Create image.
		x, y := maxFocusPos(focusMap)
		w := (x+2)*gui.FocusSpacing.X + spacingX + 17
		h := (y+1)*gui.FocusSpacing.Y + spacingY

		img := image.NewRGBA(image.Rectangle{image.ZP, image.Point{w, h}})
		draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.ZP, draw.Src)
		pBar.SetValue(pBar.Value + 0.1/i)

		// Init fonts.
		font, err = initFont(gui.Name.Font)
		if err != nil {
			showError(err)
			return
		}
		fontTreeTitle, err = initFont(gui.NationalFocusTitle.Font)
		if err != nil {
			showError(err)
			return
		}
		pBar.SetValue(pBar.Value + 0.1/i)

		if !isLineRenderingOff {
			// Draw focus tree lines.
			renderLines(img)
			pBar.SetValue(pBar.Value + 0.1/i)

			// Draw exclusivity lines.
			err = renderExclusiveLines(img)
			if err != nil {
				showError(err)
				return
			}
		}
		pBar.SetValue(pBar.Value + 0.1/i)

		// Draw focus icons.
		var focusErrMap = make(map[string]bool)
		for _, f := range focusMap {
			err = renderFocus(img, f.X*gui.FocusSpacing.X+spacingX, f.Y*gui.FocusSpacing.Y+spacingY, f.ID)
			// Save all focus icons errors into a map.
			if err != nil {
				focusErrMap[err.Error()] = true
			}
		}

		// Print out all of the errors at once, popup window on the last one.
		focusErrMapI := 0
		for errString := range focusErrMap {
			if focusErrMapI == len(focusErrMap)-1 {
				showError(errors.New(errString))
				return
			}
			ansi.Println("\x1b[31;1m" + errString + "\x1b[0m")
			focusErrMapI++
		}
		pBar.SetValue(pBar.Value + 0.1/i)

		// Save image as PNG.
		out, err := os.Create(filepath.Join(binPath, focusTreeName+".png"))
		if err != nil {
			showError(err)
			return
		}
		err = png.Encode(out, img)
		if err != nil {
			showError(err)
			return
		}
		err = out.Close()
		if err != nil {
			showError(err)
			return
		}
		ansi.Println("Image saved at \"" + filepath.Join(binPath, focusTreeName+".png") + "\"")
		pBar.SetValue(1)

		// Clear maps.
		focusMap = make(map[string]Focus)
		gfxMap = make(map[string]SpriteType)
		fontMap = make(map[string]BitmapFont)
		locMap = make(map[string]map[string]Localisation)

		// Hide progress bar.
		pBar.Hide()
		pBar.SetValue(0)
	}

	// Print out elapsed time.
	elapsedTime := time.Since(startTime)
	ansi.Printf("\x1b[30;1m"+"Elapsed time: %s\n\n"+"\x1b[0m", elapsedTime)
	running = false
}
