# diceSpy

Roll20 dice rolls transfer to OBS

# Usage (on you own risk)

* Download archive from [releases](https://github.com/averrin/diceSpy/releases)
* Run chrome with flag `--allow-running-insecure-content`
* Install and activate on roll20.net this [extension](https://chrome.google.com/webstore/detail/disable-content-security/ieelmcmcagommplceebfedjlakkhpden)
* Run diceSpy.exe
* On roll20 open WebInspector Console (`f12`)
* Type `$.getScript('http://127.0.0.1:1323/')` and press enter
* Profit. Now your (and other players) last roll(s) will be recorded in txt file near exe
* Profit #2. Now your (and other players) last roll(s) will be rendered as image file near exe
* You can use this files as OBS source for displaying your party rolls
