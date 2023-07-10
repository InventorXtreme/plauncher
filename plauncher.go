package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/andygrunwald/vdf"
)

type game struct {
	id                string
	libpath           string
	name              string
	compatpath        string
	installfoldername string
}

type steaminstance struct {
	install string
	proton  string
}

// protonrunable will be formated into the needed string by using the following spec:
type protonrunable struct {
	steamcompatclientinstallpath string //i.e.: ~/.steam/root
	steamcompatdatapath          string //i.e.: /home/inventorx/Hardspace.Shipbreaker.v1.3.0/Hardspace.Shipbreaker.v1.3.0
	protonpath                   string //i.e.: /home/inventorx/.steam/steam/steamapps/common/Proton - Experimental/proton
	exepath                      string //i.e.: /home/inventorx/Hardspace.Shipbreaker.v1.3.0/Hardspace.Shipbreaker.v1.3.0/Shipbreaker.exe
	exeargs                      string //i.e.: -no-splash
}

type proton struct {
	id                string
	libpath           string
	name              string
	installfoldername string
}

func (p protonrunable) ConvertToCall() string {
	return fmt.Sprintf("STEAM_COMPAT_CLIENT_INSTALL_PATH=\"%s\" STEAM_COMPAT_DATA_PATH=\"%s\" %s run \"%s\" %s", p.steamcompatclientinstallpath, p.steamcompatdatapath, p.protonpath, p.exepath, p.exeargs)
}

func MakeProtonRunable(s steaminstance, compatpath string, protonu proton, exepath string, exeargs string) protonrunable {
	//envargs done
	steamcompatclientinstallpath := s.install
	steamcompatdatapath := ""
	if compatpath == "" {
		b, _ := os.UserHomeDir()
		path := b + "/.plauncher/default"
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
		steamcompatdatapath = path
	}
	protonpath := protonu.installfoldername
	//exepath done
	//exeargs done
	p := protonrunable{steamcompatclientinstallpath: steamcompatclientinstallpath, steamcompatdatapath: steamcompatdatapath, protonpath: protonpath, exepath: exepath, exeargs: exeargs}
	return p
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
func GetInstallNameFromId(id string, path string) (string, error) {
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

	name, err := GetValFrom(GetMapFrom(m, "AppState"), "installdir")
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

func IsGameProton(name string) bool {
	if strings.Contains(name, "Proton") {
		return true
	} else {
		return false
	}
}

// func GetCompatDataPathForGame(g game, i steaminstance) string {

// }

func (g game) GetRunCommand(steam steaminstance, exefile string) string {
	Steamclientstring := "STEAM_COMPAT_CLIENT_INSTALL_PATH=\"" + steam.install + "\""
	return Steamclientstring
}

func GetSteamGameList(s steaminstance) ([]game, []proton) {
	var gamelist []game
	var protonlist []proton
	fmt.Println("test")
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	fpath := dirname + "/.steam/steam/steamapps/libraryfolders.vdf"
	fpath = s.install + "/steamapps/libraryfolders.vdf"
	fmt.Println(fpath)
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

		currentlibrary_applist := GetMapFrom(currentlibrary, "apps")

		for g, _ := range currentlibrary_applist {
			gameid := g
			gamelib := location
			gamename, err := GetNameFromId(gameid, gamelib)
			if err != nil {
				if err.Error() != "FileRead Err" {
					panic(err)
				} else {
					fmt.Println("Missing Library @ " + location + " @ Lookup of " + gameid)
				}
			}
			compatpath := location + "/steamapps/compatdata/" + string(gameid)
			installpathdir, err := GetInstallNameFromId(gameid, gamelib)
			installpath := location + "/steamapps/common/" + installpathdir
			if err == nil {
				e, _ := IsGameWindows(gameid, gamelib)
				if e {
					gamelist = append(gamelist, game{gameid, gamelib, gamename, compatpath, installpath})

				} else if IsGameProton(gamename) {
					protonlist = append(protonlist, proton{gameid, gamelib, gamename, installpath})
				}

			} else {
				if err.Error() != "FileRead Err" {
					panic(err)
				} else {
					fmt.Println("Missing Library @ " + location + " @ Lookup of " + gameid)
				}

			}

		}

	}

	return gamelist, protonlist
}

// func GetListOfProtons

func main() {
	b, _ := os.UserHomeDir()
	i := steaminstance{install: (b + "/.steam/steam/")}
	a, p := GetSteamGameList(i)

	//fmt.Println(a)
	BetterPrintGameList(a)
	BetterPrintProtonList(p)
	protonid := 0
	for id, k := range p {
		if k.name == "Proton Experimental" {
			protonid = id
		}
	}
	e := exec.Command("bash", "-c", "echo $sneed")
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	e.Env = os.Environ()
	e.Stdout = mw
	e.Stderr = mw
	e.Run()
	ex := MakeProtonRunable(i, "", p[protonid], "~/h.exe", "")
	fmt.Println(ex)
	fmt.Println(ex.ConvertToCall())
}
