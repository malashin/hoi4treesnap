package main

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

var pdxRule = `
	entry                = '' scopeBody$;

	declr                = lval '=' rval [';'] [@comment];
	declrScope           = lval '=' scope [';'] [@comment];
	comparison           = lval @operators rval [';'] [@comment];
	list                 = @anyType {@anyType} [';']  [@comment];

	lval                 = @date|@int|@var|'"'#@string#'"';
	rval                 = @date|@hex|@percent|@var|@number|'"'#@string#'"';

	scope                = '{' (scopeBody|@empty) ('}'|empty);
	scopeBody            = (@declr|@declrScope|@comparison|@list){@declr|@declrScope|@comparison|@list};
	comment              = '#'#{#!\x0a#!\x0d#!$#anyRune};

	int                  = ['-']digit#{#digit};
	float                = ['-'][int]#'.'#int;
	number               = float|int;
	percent              = int#'%'#'%';
	string               = {!'"'#stringChar};
	var                  = symbol#{#symbol};
	date                 = int#'.'#int#'.'#int#['.'#int];
	bool                 = 'yes'|'no';
	hex                  = '0x'#(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter)(digit|letter);
	anyType              = number|percent|'"'#string#'"'|var|date|bool|hex;

	                     = {spaces|@comment};
	spaces               = \x00..\x20;
	anyRune              = \x00..$;
	digit                = '0'..'9';
	letter               = 'a'..'z'|'A'..'Z'|'а'..'я'|'А'..'Я'|\u00c0..\u00d6|\u00d8..\u00f6|\u00f8..\u00ff|\u0100..\u017f|\u0180..\u024f|\u0400..\u04ff|\u0500..\u052f;
	operators            = '<'|'>';
	symbol               = digit|letter|'_'|':'|'@'|'.'|'-'|'^'|\u0027;
	stringChar           = ('\"'|anyRune);
	empty                = '';
`

// pair                 = @key ':' [@number] '"'#@value#'"' [@comment];

var ymlRule = `
	entry                = @language#':' {@pair};

	pair                 = @key ':' [@number] @value [@comment];
	comment              = '#'#{#!\x0a#!\x0d#!$#anyRune};

	language             = 'l_'#symbol#{#symbol};
	key                  = symbol#{#symbol};
	number               = digit#{#digit};
	value                = '"'#{#anyRune};

	                     = {spaces|@comment};
	spaces               = \x00..\x20;
	anyRune              = \x00..\x09|\x0b..\x0c|\x0e..$;
	digit                = '0'..'9';
	letter               = 'a'..'z'|'A'..'Z'|'а'..'я'|'А'..'Я'|\u00c0..\u00d6|\u00d8..\u00f6|\u00f8..\u00ff|\u0100..\u017f|\u0180..\u024f|\u0400..\u04ff|\u0500..\u052f;
	symbol               = digit|letter|'_'|'@'|'.'|'-';
	empty                = '';
`

func parseFocus(path string) error {
	fmt.Println(path)
	f, err := readFile(path)
	if err != nil {
		return err
	}

	if len(f) > 0 {
		// Remove utf-8 bom if found.
		if bytes.HasPrefix([]byte(f), utf8bom) {
			f = string(bytes.TrimPrefix([]byte(f), utf8bom))
		}

		node, err := pdx.Parse(f)
		if err != nil {
			return err
		}
		_ = node
		// fmt.Println(ptool.TreeToString(node, pdx.ByID))
		err = traverseFocus(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func traverseFocus(root *ptool.TNode) error {
	for _, node := range root.Links {
		nodeType := pdx.ByID(node.Type)
		switch nodeType {
		case "declrScope":
			switch strings.ToLower(node.Links[0].Value) {
			case "focus", "shared_focus":
				var f Focus
				f.AllowBranch = true
				f.Available = true
				var err error
				var n float64
				for _, link := range node.Links {
					nodeType := pdx.ByID(link.Type)
					switch nodeType {
					case "declr":
						switch strings.ToLower(link.Links[0].Value) {
						case "id":
							f.ID = link.Links[1].Value
							locList = append(locList, link.Links[1].Value)
						case "icon":
							f.Icon = link.Links[1].Value
							gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
						case "text":
							f.Text = link.Links[1].Value
							locList = append(locList, link.Links[1].Value)
						case "x":
							n, err = strconv.ParseFloat(link.Links[1].Value, 64)
							if err != nil {
								return err
							}
							f.X = int(math.Trunc(n))
						case "y":
							n, err = strconv.ParseFloat(link.Links[1].Value, 64)
							if err != nil {
								return err
							}
							f.Y = int(math.Trunc(n))
						case "relative_position_id":
							f.RelativePositionID = link.Links[1].Value
						}
					case "declrScope":
						switch strings.ToLower(link.Links[0].Value) {
						case "prerequisite":
							var p []string
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "declr":
									switch strings.ToLower(link.Links[0].Value) {
									case "focus":
										p = append(p, link.Links[1].Value)
									}
								}
							}
							f.Prerequisite = append(f.Prerequisite, p)
						case "mutually_exclusive":
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "declr":
									switch strings.ToLower(link.Links[0].Value) {
									case "focus":
										f.MutuallyExclusive = append(f.MutuallyExclusive, link.Links[1].Value)
									}
								}
							}
						case "allow_branch":
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "declr":
									switch strings.ToLower(link.Links[0].Value) {
									case "always":
										if strings.ToLower(link.Links[1].Value) == "no" {
											f.AllowBranch = false
										}
									case "has_country_flag":
										if strings.ToLower(link.Links[1].Value) == "romanov_enabled" { // Poland tree workaround
											f.AllowBranch = false
										}
									}
								case "declrScope":
									switch strings.ToLower(link.Links[0].Value) {
									case "not":
										for _, link := range link.Links {
											nodeType := pdx.ByID(link.Type)
											switch nodeType {
											case "declr":
												switch strings.ToLower(link.Links[0].Value) {
												case "has_dlc":
													f.AllowBranch = false
												}
											}
										}
									}
								}
							}
						case "available":
							for _, link := range link.Links {
								if len(link.Links) > 0 {
									f.Available = false
								}
							}
						}
					}
				}
				focusMap[f.ID] = f
			default:
				err := traverseFocus(node)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseGUI(path string) error {
	fPath := filepath.Join(path, "interface", "nationalfocusview.gui")
	fmt.Println(fPath)

	f, err := readFile(fPath)
	if err != nil {
		return err
	}

	if len(f) > 0 {
		// Remove utf-8 bom if found.
		if bytes.HasPrefix([]byte(f), utf8bom) {
			f = string(bytes.TrimPrefix([]byte(f), utf8bom))
		}

		node, err := pdx.Parse(f)
		if err != nil {
			return err
		}
		_ = node
		// fmt.Println(ptool.TreeToString(node, pdx.ByID))
		err = traverseGUI(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func traverseGUI(root *ptool.TNode) error {
	var err error
	for _, node := range root.Links {
		nodeType := pdx.ByID(node.Type)
		switch nodeType {
		case "declrScope":
			switch strings.ToLower(node.Links[0].Value) {
			case "containerwindowtype":
				nfv := false
				nfi := false
				nfl := false
				nfei := false
				for _, link := range node.Links {
					if pdx.ByID(link.Type) == "declr" {
						if strings.ToLower(link.Links[0].Value) == "name" && link.Links[1].Value == "nationalfocusview" {
							nfv = true
						}
						if strings.ToLower(link.Links[0].Value) == "name" && link.Links[1].Value == "national_focus_item" {
							nfi = true
						}
						if strings.ToLower(link.Links[0].Value) == "name" && link.Links[1].Value == "national_focus_link" {
							nfl = true
						}
						if strings.ToLower(link.Links[0].Value) == "name" && link.Links[1].Value == "national_focus_exclusive_item" {
							nfei = true
						}
					}
				}

				switch {
				case nfv:
					for _, link := range node.Links {
						if len(link.Links) > 0 {
							switch strings.ToLower(link.Links[0].Value) {
							case "instanttextboxtype":
								var t InstantTextboxType
								for _, link := range link.Links {
									if len(link.Links) > 0 {
										switch strings.ToLower(link.Links[0].Value) {
										case "name":
											t.Name = link.Links[1].Value
										case "position":
											for _, link := range link.Links {
												nodeType := pdx.ByID(link.Type)
												switch nodeType {
												case "declr":
													switch strings.ToLower(link.Links[0].Value) {
													case "x":
														t.Position.X, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													case "y":
														t.Position.Y, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													}
												}
											}
										case "font":
											t.Font = link.Links[1].Value
											gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
										case "text":
											t.Text = link.Links[1].Value
										case "maxwidth":
											t.MaxWidth, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "maxheight":
											t.MaxHeight, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "format":
											t.Format = link.Links[1].Value
										case "vertical_alignment":
											t.VerticalAlignment = link.Links[1].Value
										}
									}
								}
								if t.Name == "national_focus_title" {
									gui.NationalFocusTitle = t
								}
							}
						}
					}

				case nfi:
					for _, link := range node.Links {
						if len(link.Links) > 0 {
							switch strings.ToLower(link.Links[0].Value) {
							case "name":
								gui.NationalFocusItem.Name = link.Links[1].Value
							case "position":
								for _, link := range link.Links {
									nodeType := pdx.ByID(link.Type)
									switch nodeType {
									case "declr":
										switch strings.ToLower(link.Links[0].Value) {
										case "x":
											gui.NationalFocusItem.Position.X, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "y":
											gui.NationalFocusItem.Position.Y, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
							case "size":
								for _, link := range link.Links {
									nodeType := pdx.ByID(link.Type)
									switch nodeType {
									case "declr":
										switch strings.ToLower(link.Links[0].Value) {
										case "width":
											gui.NationalFocusItem.Width, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "height":
											gui.NationalFocusItem.Height, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
							case "buttontype":
								var button ButtonType
								for _, link := range link.Links {
									if len(link.Links) > 0 {
										switch strings.ToLower(link.Links[0].Value) {
										case "name":
											button.Name = link.Links[1].Value
										case "position":
											for _, link := range link.Links {
												nodeType := pdx.ByID(link.Type)
												switch nodeType {
												case "declr":
													switch strings.ToLower(link.Links[0].Value) {
													case "x":
														button.Position.X, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													case "y":
														button.Position.Y, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													}
												}
											}
										case "spritetype":
											button.SpriteType = link.Links[1].Value
											gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
										case "quadtexturesprite":
											button.SpriteType = link.Links[1].Value
											gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
										case "centerposition":
											button.CenterPosition = link.Links[1].Value
										case "orientation":
											button.Orientation = link.Links[1].Value
										}
									}
								}
								switch strings.ToLower(button.Name) {
								case "bg":
									gui.BG = button
								case "symbol":
									gui.Symbol = button
								}
							case "instanttextboxtype":
								name := false
								for _, link := range link.Links {
									if pdx.ByID(link.Type) == "declr" {
										if strings.ToLower(link.Links[0].Value) == "name" && link.Links[1].Value == "name" {
											name = true
										}
									}
								}
								if name {
									for _, link := range link.Links {
										if len(link.Links) > 0 {
											switch strings.ToLower(link.Links[0].Value) {
											case "name":
												gui.Name.Name = link.Links[1].Value
											case "position":
												for _, link := range link.Links {
													nodeType := pdx.ByID(link.Type)
													switch nodeType {
													case "declr":
														switch strings.ToLower(link.Links[0].Value) {
														case "x":
															gui.Name.Position.X, err = strconv.Atoi(link.Links[1].Value)
															if err != nil {
																return err
															}
														case "y":
															gui.Name.Position.Y, err = strconv.Atoi(link.Links[1].Value)
															if err != nil {
																return err
															}
														}
													}
												}
											case "font":
												gui.Name.Font = link.Links[1].Value
												gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
											case "text":
												gui.Name.Text = link.Links[1].Value
											case "maxwidth":
												gui.Name.MaxWidth, err = strconv.Atoi(link.Links[1].Value)
												if err != nil {
													return err
												}
											case "maxheight":
												gui.Name.MaxHeight, err = strconv.Atoi(link.Links[1].Value)
												if err != nil {
													return err
												}
											case "format":
												gui.Name.Format = link.Links[1].Value
											case "vertical_alignment":
												gui.Name.VerticalAlignment = link.Links[1].Value
											}
										}
									}
								}
							}
						}
					}

				case nfl:
					for _, link := range node.Links {
						if len(link.Links) > 0 {
							switch strings.ToLower(link.Links[0].Value) {
							case "name":
								gui.NationalFocusLink.Name = link.Links[1].Value
							case "position":
								for _, link := range link.Links {
									nodeType := pdx.ByID(link.Type)
									switch nodeType {
									case "declr":
										switch strings.ToLower(link.Links[0].Value) {
										case "x":
											gui.NationalFocusLink.Position.X, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "y":
											gui.NationalFocusLink.Position.Y, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
							case "size":
								for _, link := range link.Links {
									nodeType := pdx.ByID(link.Type)
									switch nodeType {
									case "declr":
										switch strings.ToLower(link.Links[0].Value) {
										case "width":
											gui.NationalFocusLink.Width, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "height":
											gui.NationalFocusLink.Height, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
							case "icontype":
								var icon IconType
								for _, link := range link.Links {
									if len(link.Links) > 0 {
										switch strings.ToLower(link.Links[0].Value) {
										case "name":
											icon.Name = link.Links[1].Value
										case "position":
											for _, link := range link.Links {
												nodeType := pdx.ByID(link.Type)
												switch nodeType {
												case "declr":
													switch strings.ToLower(link.Links[0].Value) {
													case "x":
														icon.Position.X, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													case "y":
														icon.Position.Y, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													}
												}
											}
										case "spritetype":
											icon.SpriteType = link.Links[1].Value
											gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
										case "frame":
											icon.Frame, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
								if strings.ToLower(icon.Name) == "link" {
									gui.Link = icon
								}
							}
						}
					}

				case nfei:
					for _, link := range node.Links {
						if len(link.Links) > 0 {
							switch strings.ToLower(link.Links[0].Value) {
							case "name":
								gui.NationalFocusExclusiveItem.Name = link.Links[1].Value
							case "position":
								for _, link := range link.Links {
									nodeType := pdx.ByID(link.Type)
									switch nodeType {
									case "declr":
										switch strings.ToLower(link.Links[0].Value) {
										case "x":
											gui.NationalFocusExclusiveItem.Position.X, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "y":
											gui.NationalFocusExclusiveItem.Position.Y, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
							case "size":
								for _, link := range link.Links {
									nodeType := pdx.ByID(link.Type)
									switch nodeType {
									case "declr":
										switch strings.ToLower(link.Links[0].Value) {
										case "width":
											gui.NationalFocusExclusiveItem.Width, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										case "height":
											gui.NationalFocusExclusiveItem.Height, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
							case "icontype":
								var icon IconType
								for _, link := range link.Links {
									if len(link.Links) > 0 {
										switch strings.ToLower(link.Links[0].Value) {
										case "name":
											icon.Name = link.Links[1].Value
										case "position":
											for _, link := range link.Links {
												nodeType := pdx.ByID(link.Type)
												switch nodeType {
												case "declr":
													switch strings.ToLower(link.Links[0].Value) {
													case "x":
														icon.Position.X, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													case "y":
														icon.Position.Y, err = strconv.Atoi(link.Links[1].Value)
														if err != nil {
															return err
														}
													}
												}
											}
										case "spritetype":
											icon.SpriteType = link.Links[1].Value
											gfxList = append(gfxList, "\""+link.Links[1].Value+"\"")
										case "frame":
											icon.Frame, err = strconv.Atoi(link.Links[1].Value)
											if err != nil {
												return err
											}
										}
									}
								}
								switch strings.ToLower(icon.Name) {
								case "link1":
									gui.Link1 = icon
								case "link2":
									gui.Link2 = icon
								case "left":
									gui.Left = icon
								case "right":
									gui.Right = icon
								case "mid":
									gui.Mid = icon
								}
							}
						}
					}
				}

			case "positiontype":
				var name string
				var pos image.Point
				for _, link := range node.Links {
					if len(link.Links) > 0 {
						switch strings.ToLower(link.Links[0].Value) {
						case "name":
							name = link.Links[1].Value
						case "position":
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "declr":
									switch strings.ToLower(link.Links[0].Value) {
									case "x":
										pos.X, err = strconv.Atoi(link.Links[1].Value)
										if err != nil {
											return err
										}
									case "y":
										pos.Y, err = strconv.Atoi(link.Links[1].Value)
										if err != nil {
											return err
										}
									}
								}
							}
						}
					}
				}
				switch strings.ToLower(name) {
				case "focus_spacing":
					gui.FocusSpacing = pos
				case "link_spacing":
					gui.LinkSpacing = pos
				case "link_offsets":
					gui.LinkOffsets = pos
				case "link_begin":
					gui.LinkBegin = pos
				case "link_end":
					gui.LinkEnd = pos
				case "exclusive_offset":
					gui.ExclusiveOffset = pos
				case "exclusive_offset_left":
					gui.ExclusiveOffsetLeft = pos
				case "exclusive_positioning":
					gui.ExclusivePositioning = pos
				}

			default:
				err := traverseGUI(node)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseGFX(path string, i int) error {
	gfxFiles, err := WalkMatchExt(filepath.Join(path, "interface"), ".gfx")
	if err != nil {
		return err
	}
	for _, fPath := range gfxFiles {
		f, err := readFile(fPath)
		if err != nil {
			return err
		}

		if stringContainsSlice(f, gfxList) {
			fmt.Println(fPath)
			if len(f) > 0 {
				// Remove utf-8 bom if found.
				if bytes.HasPrefix([]byte(f), utf8bom) {
					f = string(bytes.TrimPrefix([]byte(f), utf8bom))
				}

				node, err := pdx.Parse(f)
				if err != nil {
					return err
				}
				_ = node
				// fmt.Println(ptool.TreeToString(node, pdx.ByID))
				err = traverseGFX(node, path)
				if err != nil {
					return err
				}
			}
		}
		pBar.SetValue(pBar.Value + 0.4/float64(i)/float64(len(gfxFiles)))
	}
	return nil
}

func traverseGFX(root *ptool.TNode, path string) error {
	var err error
	for _, node := range root.Links {
		nodeType := pdx.ByID(node.Type)
		switch nodeType {
		case "declrScope":
			switch strings.ToLower(node.Links[0].Value) {
			case "spritetype", "corneredtilespritetype":
				var s SpriteType
				for _, link := range node.Links {
					nodeType := pdx.ByID(link.Type)
					switch nodeType {
					case "declr":
						switch strings.ToLower(link.Links[0].Value) {
						case "name":
							s.Name = link.Links[1].Value
						case "texturefile":
							s.TextureFile = filepath.Join(path, link.Links[1].Value)
						case "noofframes":
							s.NoOfFrames, err = strconv.Atoi(link.Links[1].Value)
							if err != nil {
								return err
							}
						}
					}
				}
				gfxMap[s.Name] = s
			case "bitmapfont":
				var b BitmapFont
				for _, link := range node.Links {
					nodeType := pdx.ByID(link.Type)
					switch nodeType {
					case "declr":
						switch strings.ToLower(link.Links[0].Value) {
						case "name":
							b.Name = link.Links[1].Value
						case "path":
							b.Path = filepath.Join(path, link.Links[1].Value)
						}
					case "declrScope":
						switch strings.ToLower(link.Links[0].Value) {
						case "fontfiles":
							for _, link := range link.Links {
								nodeType := pdx.ByID(link.Type)
								switch nodeType {
								case "list":
									for _, link := range link.Links {
										nodeType := pdx.ByID(link.Type)
										switch nodeType {
										case "anyType":
											b.Fontfiles = append(b.Fontfiles, filepath.Join(path, trimQuotes(link.Value)))
										}

									}
								}
							}
						}
					}
				}
				if len(b.Fontfiles) < 1 && b.Path != "" {
					b.Fontfiles = append(b.Fontfiles, b.Path)
				}
				fontMap[b.Name] = b
			default:
				err = traverseGFX(node, path)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseLoc(path string, i int) error {
	locFiles, err := WalkMatchExt(filepath.Join(path, "localisation"), ".yml")
	if err != nil {
		return err
	}

	var locReplaceFiles []string
	if _, err := os.Stat(filepath.Join(path, "localisation", "replace")); os.IsExist(err) {
		locReplaceFiles, err = WalkMatchExt(filepath.Join(path, "localisation", "replace"), ".yml")
		if err != nil {
			return err
		}
	}

	locFiles = append(locFiles, locReplaceFiles...)

	for _, lPath := range locFiles {
		f, err := readFile(lPath)
		if err != nil {
			return err
		}
		if stringContainsSlice(f, locList) {
			if len(f) > 0 {
				// Remove utf-8 bom if found.
				if bytes.HasPrefix([]byte(f), utf8bom) {
					f = string(bytes.TrimPrefix([]byte(f), utf8bom))
				}

				// Skip file if it contains a wrong language.
				if !strings.HasPrefix(strings.TrimSpace(f), language) {
					continue
				}

				fmt.Println(lPath)

				node, err := yml.Parse(f)
				if err != nil {
					return err
				}
				_ = node
				// fmt.Println(ptool.TreeToString(node, yml.ByID))

				err = traverseLoc(node)
				if err != nil {
					return err
				}
			}
		}
		pBar.SetValue(pBar.Value + 0.4/float64(i)/float64(len(locFiles)))
	}
	return nil
}

func traverseLoc(root *ptool.TNode) error {
	lang := "l_english"
	for _, node := range root.Links {
		nodeType := yml.ByID(node.Type)
		switch nodeType {
		case "language":
			lang = node.Value
			if _, ok := locMap[lang]; !ok {
				locMap[lang] = make(map[string]Localisation)
			}
		case "pair":
			var l Localisation
			for _, link := range node.Links {
				nodeType := yml.ByID(link.Type)
				switch nodeType {
				case "key":
					l.Key = link.Value
				case "number":
					l.Number = link.Value
				case "value":
					l.Value = trimQuotes(link.Value)
				}
			}
			locMap[lang][l.Key] = l
		default:
			err := traverseLoc(node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
