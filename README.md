__hoi4treesnap__ generates Hearts of Iron IV focus tree screenshots.

### How to use:
1. Select focus tree file from "/common/national_focus".
2. Select Hearts of Iron IV game folder. It will be saved for later use after the first time.
3. If you need other mods, dependancies for example, select those.
4. Press "Generate image". Output will be saved next to the hoi4treesnap binary.

### Possible issues:
* The file parser is stricter then PDX one, so you might need to fix those errors if they are reported.
* DDS decoder can only read RGBA 8.8.8.8 images, not compressed ones like DXT1 or DXT5, so you will need to resave them as RGBA.

### Known issues:
* You can't generate single image for shared focus trees. You'll have to combine them from seperate images.
* There is no country name in the image. Might be added later either through parsing of the files or just asking the user to input the name.
* If focus title uses scripted localisation, it will be rendered as a scripted localisation string instead of the apropriate name. Might ask user to enter apropriate titles if those are found later on.

<img src="https://i.imgur.com/1Wepd3Z.png">
<img src="https://i.imgur.com/MKPV5Cc.png">
