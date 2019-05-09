__hoi4treesnap__ generates Hearts of Iron IV focus tree screenshots.

The tool itself does not contain any textures and picks them up from the HOI4 base game or a mod that contains selected focus trees. That includes all focus tree graphics: focus icons, focus tree plaques, focus tree lines and fonts. `nationalfocusview.gui` is being parsed to pick on your changes to it, so the output image looks quite similar to what you see in the game, even a modded one.

### How to use:
1. Select focus tree file from `/common/national_focus`.
2. Select Hearts of Iron IV game folder. It will be saved for later use after the first time.
3. If you need other mods, dependencies for example, select those.
4. If you want to use non-english localisation press `Select localisation language`.
5. Press `Generate image`. Output will be saved next to the hoi4treesnap binary.

### Possible issues:
* The file parser is stricter then PDX one, so you might need to fix those errors if they are reported.
* DDS decoder can only read RGBA 8.8.8.8 images, not compressed ones like DXT1 or DXT5, so you will need to resave them as RGBA.

### Known issues:
* You can't generate single image for shared focus trees. You'll have to combine them from separate images.
* There is no country name in the image. Might be added later either through parsing of the files or just asking the user to input the name.
* If focus title uses scripted localization, it will be rendered as a scripted localization string instead of the appropriate name. Might ask user to enter appropriate titles if those are found later on.

### Menu:
<img src="https://i.imgur.com/84sotcl.png">

### Output examples:
<img src="https://i.imgur.com/MKPV5Cc.png">
<img src="https://i.imgur.com/8Bq71l1.png">
