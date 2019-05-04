package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
	"github.com/malashin/bmfonter"
)

func nodesToString(node *ptool.TNode) []string {
	s := []string{}
	for _, n := range node.Links {
		s = append(s, nodesToString(n)...)
	}
	if node.Value != "" {
		s = append(s, node.Value)
	}
	return s
}

func readFile(path string) (string, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func encodeCacheFile(i interface{}, path string) error {
	err := e.Encode(i)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, buf.Bytes(), 0755)
	if err != nil {
		return err
	}
	buf.Reset()
	return nil
}

func decodeCacheFile(i interface{}, path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = buf.Write(b)
	if err != nil {
		return err
	}
	err = d.Decode(i)
	if err != nil {
		return err
	}
	buf.Reset()
	return nil
}

func fillAbsoluteFocusPositions(finished bool) bool {
	for _, f1 := range focusMap {
		if f1.RelativePositionID == "" {
			for _, f2 := range focusMap {
				if f2.RelativePositionID == f1.ID {
					f2.X += focusMap[f1.ID].X
					f2.Y += focusMap[f1.ID].Y
					f2.RelativePositionID = ""
					focusMap[f2.ID] = f2
					finished = false
				}
			}
		}
	}
	if !finished {
		finished = fillAbsoluteFocusPositions(true)
	}
	return finished
}

// buildFocusTree adds children data to each focus.
// Sorts children by X coordinate from left to right.
func fillFocusChildAndParentData() {
	for _, c := range focusMap {
		for _, g := range c.Prerequisite {
			solid := true
			if len(g) > 1 {
				solid = false
			}
			for _, f := range g {
				p := focusMap[f]
				p.Children = append(p.Children, Child{c.ID, solid})
				focusMap[p.ID] = p
			}
		}
	}

	for _, f := range focusMap {
		sort.Slice(f.Children, func(i, j int) bool { return focusMap[f.Children[i].ID].X < focusMap[f.Children[j].ID].X })
		focusMap[f.ID] = f
		fillAllowBranchData(f)
	}

	for _, p := range focusMap {
		for i, child := range p.Children {

			c := focusMap[child.ID]
			if c.In == nil {
				m := make(map[int]FocusLine)
				c.In = m
			}
			if !c.AllowBranch {
				continue
			}
			a := c.In[p.Y]
			if child.Solid {
				a.Set(S)
				p.Out.Set(S)
			} else {
				for _, child2 := range p.Children {
					c2 := focusMap[child2.ID]
					switch {
					case c.X < p.X && c2.X < c.X && child2.Solid:
						a.Set(S)
					case c.X > p.X && c2.X > c.X && child2.Solid:
						a.Set(S)
					case c.X == p.X && child2.Solid:
						a.Set(S)
					}
				}
			}

			switch {
			case c.X < p.X:
				a.Set(D | R)
				if i != 0 {
					a.Set(L)
				}
				p.Out.Set(U | L)
			case c.X == p.X:
				a.Set(U | D)
				if i > 0 && focusMap[p.Children[i-1].ID].AllowBranch {
					a.Set(L)
				}
				if i != len(p.Children)-1 && focusMap[p.Children[i+1].ID].AllowBranch {
					a.Set(R)
				}
				p.Out.Set(U | D)
			case c.X > p.X:
				a.Set(D | L)
				if i != len(p.Children)-1 {
					a.Set(R)
				}
				p.Out.Set(U | R)
			}

			for _, pSlice := range c.Prerequisite {
				for _, p2 := range pSlice {
					p2 := focusMap[p2]
					if p.Y > p2.Y {
						a.Set(U)
					}
				}
			}

			c.In[p.Y] = a
			focusMap[c.ID] = c
		}
		focusMap[p.ID] = p
	}
}

func (l *FocusLine) Set(d Dir) {
	l.Dir |= d
}

func (l *FocusLine) Get() image.Image {
	switch l.Dir {
	case 3:
		return UDdash
	case 5:
		return ULdash
	case 6:
		return DLdash
	case 7:
		return UDLdash
	case 9:
		return URdash
	case 10:
		return DRdash
	case 11:
		return UDRdash
	case 12:
		return LRdash
	case 13:
		return ULRdash
	case 14:
		return DLRdash
	case 15:
		return UDLRdash

	case 19:
		return UD
	case 21:
		return UL
	case 22:
		return DL
	case 23:
		return UDL
	case 25:
		return UR
	case 26:
		return DR
	case 27:
		return UDR
	case 28:
		return LR
	case 29:
		return ULR
	case 30:
		return DLR
	case 31:
		return UDLR
	}
	return nil
}

func containsInt(s []int, a int) bool {
	for _, b := range s {
		if a == b {
			return true
		}
	}
	return false
}

func containsString(s []string, a string) bool {
	for _, b := range s {
		if a == b {
			return true
		}
	}
	return false
}

func containsPoint(s []image.Point, a image.Point) bool {
	for _, b := range s {
		if a == b {
			return true
		}
	}
	return false
}

// maxFocusPos returns maximum x and y values in focus tree.
func maxFocusPos(m map[string]Focus) (x, y int) {
	for _, f := range m {
		if f.X > x {
			x = f.X
		}
		if f.Y > y {
			y = f.Y
		}
	}
	return
}

func fillAllowBranchData(f Focus) {
	if !f.AllowBranch {
		for _, child := range f.Children {
			c := focusMap[child.ID]
			c.AllowBranch = false
			focusMap[child.ID] = c
			fillAllowBranchData(c)
		}
	}
}

func maxYinRange(m map[int]FocusLine, y int) int {
	var max int
	for i := range m {
		if i > max && i > y {
			max = i
		}
	}
	return max
}

func stringContainsSlice(s string, slice []string) bool {
	for _, substr := range slice {
		c := strings.Contains(s, substr)
		if c {
			return true
		}
	}
	return false
}

func useModsTexturesIfPresent() {
	if len(modPaths) <= 1 {
		return
	}
	for k, v := range gfxMap {
		if strings.HasPrefix(v.TextureFile, gamePath) {
			gfx := strings.TrimPrefix(v.TextureFile, gamePath)
			for _, p := range modPaths[1:] {
				if _, err := os.Stat(filepath.Join(p, gfx)); err == nil {
					v.TextureFile = filepath.Join(p, gfx)
					gfxMap[k] = v
				}
			}
		}
	}
}

func initFont(fontName string) (bmfonter.Font, error) {
	var font bmfonter.Font
	bmfont, ok := fontMap[fontName]
	if !ok {
		return font, fmt.Errorf("font \"" + fontName + "\" not found")
	}

	if len(bmfont.Fontfiles) < 1 {
		return font, fmt.Errorf("font \"" + fontName + "\" has no associated files")
	}

	// Init font.
	font, err := bmfonter.InitFont(bmfont.Fontfiles[0]+".fnt", bmfont.Fontfiles[0]+".dds")
	if err != nil {
		return font, err
	}

	if len(bmfont.Fontfiles) > 1 {
		for i := 1; i < len(bmfont.Fontfiles); i++ {
			err = font.AddSubFont(bmfont.Fontfiles[i]+".fnt", bmfont.Fontfiles[i]+".dds")
			return font, err
		}
	}

	return font, nil
}
