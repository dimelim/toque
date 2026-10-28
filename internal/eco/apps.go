package eco

var apps = map[string]string{
	"chrome.exe":                "Google Chrome",
	"msedge.exe":                "Microsoft Edge",
	"firefox.exe":               "Firefox",
	"opera.exe":                 "Opera",
	"brave.exe":                 "Brave",
	"vivaldi.exe":               "Vivaldi",
	"steam.exe":                 "Steam",
	"steamwebhelper.exe":        "Steam",
	"epicgameslauncher.exe":     "Epic Games",
	"epicwebhelper.exe":         "Epic Games",
	"battle.net.exe":            "Battle.net",
	"eadesktop.exe":             "EA app",
	"upc.exe":                   "Ubisoft Connect",
	"galaxyclient.exe":          "GOG Galaxy",
	"signalrgb.exe":             "SignalRGB",
	"wallpaper32.exe":           "Wallpaper Engine",
	"wallpaper64.exe":           "Wallpaper Engine",
	"webwallpaper32.exe":        "Wallpaper Engine",
	"lively.exe":                "Lively Wallpaper",
	"onedrive.exe":              "OneDrive",
	"googledrivefs.exe":         "Google Drive",
	"dropbox.exe":               "Dropbox",
	"creative cloud.exe":        "Adobe Creative Cloud",
	"adobe desktop service.exe": "Adobe Creative Cloud",
	"ccxprocess.exe":            "Adobe Creative Cloud",
}

func Label(exe string) (string, bool) {
	l, ok := apps[exe]
	return l, ok
}
