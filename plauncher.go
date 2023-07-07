package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/andygrunwald/vdf"
)

type game struct {
	id         string
	libpath    string
	name       string
	compatpath string
}

type steaminstance struct {
	install string
	proton  string
}

// protonrunable will be formated into the needed string by using the following spec:
type protonrunable struct {
	envargs                      string //i.e.: PROTON_ENABLE_NVAPI=1 VKD3D_CONFIG=dxr PROTON_ENABLE_NVAPI=1 VKD3D_CONFIG=dxr
	steamcompatclientinstallpath string //i.e.: ~/.steam/root
	steamcompatdatapath          string //i.e.: /home/inventorx/Hardspace.Shipbreaker.v1.3.0/Hardspace.Shipbreaker.v1.3.0
	protonpath                   string //i.e.: /home/inventorx/.steam/steam/steamapps/common/Proton - Experimental/proton
	exepath                      string //i.e.: /home/inventorx/Hardspace.Shipbreaker.v1.3.0/Hardspace.Shipbreaker.v1.3.0/Shipbreaker.exe
	exeargs                      string //i.e.: -no-splash
}

func (p protonrunable) ConvertToCall() string {
	return fmt.Sprintf("%s STEAM_COMPAT_CLIENT_INSTALL_PATH=\"%s\" STEAM_COMPAT_DATA_PATH=\"%s\" %s run \"%s\" %s", p.envargs, p.steamcompatclientinstallpath, p.steamcompatdatapath, p.protonpath, p.exepath, p.exeargs)
}

func GetMapFrom(source map[string]interface{}, key string) map[string]interface{} {
	return source[key].(map[string]interface{})
}

func GetValFrom(source map[string]interface{}, key string) (value string, e error) {
	e = nil
	defer func() {
		r := recover()
		if r != nil {

			e = errors.New("Nil Val")
			value = ""
		}
	}()
	value = source[key].(string)

	return
}

func GetNameFromId(id string, path string) (string, error) {
	filepath := path + "/steamapps/appmanifest_" + id + ".acf"

	f, err := os.Open(filepath)
	if err != nil {
		return "", errors.New("FileRead Err")
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		panic(err)

	}

	name, err := GetValFrom(GetMapFrom(m, "AppState"), "name")
	if err != nil {
		panic(err)
	}
	return name, nil

}

func IsGameWindows(id string, path string) (bool, error) {
	filepath := path + "/steamapps/appmanifest_" + id + ".acf"

	f, err := os.Open(filepath)
	if err != nil {
		return false, errors.New("FileRead Err")
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		panic(err)

	}
	platform, _ := GetValFrom(GetMapFrom(GetMapFrom(m, "AppState"), "UserConfig"), "platform_override_source")
	if platform == "windows" {
		return true, nil
	} else {
		return false, nil
	}

}

// func GetCompatDataPathForGame(g game, i steaminstance) string {

// }

func (g game) GetRunCommand(steam steaminstance, exefile string) string {
	Steamclientstring := "STEAM_COMPAT_CLIENT_INSTALL_PATH=\"" + steam.install + "\""
	return Steamclientstring
}

func GetSteamGameList() []game {
	var gamelist []game
	fmt.Println("test")
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	fpath := dirname + "/.steam/steam/steamapps/libraryfolders.vdf"
	f, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m, err := p.Parse()
	if err != nil {
		panic(err)
	}
	set := GetMapFrom(m, "libraryfolders")

	for k, _ := range set {
		currentlibrary := GetMapFrom(set, k)
		location, err := GetValFrom(currentlibrary, "path")
		if err != nil {
			panic(err)
		}
		fmt.Println(location)

		currentlibrary_applist := GetMapFrom(currentlibrary, "apps")

		for g, _ := range currentlibrary_applist {
			gameid := g
			gamelocation := location
			gamename, err := GetNameFromId(gameid, gamelocation)
			compatpath := location + "/steamapps/compatdata/" + string(gameid)
			if err == nil {
				e, _ := IsGameWindows(gameid, gamelocation)
				if e {
					gamelist = append(gamelist, game{gameid, gamelocation, gamename, compatpath})

				}

			}

		}

	}

	return gamelist
}

func main() {
	a := GetSteamGameList()
	//fmt.Println(a)
	BetterPrintGameList(a)
}
