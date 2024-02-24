package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andygrunwald/vdf"
	flag "github.com/spf13/pflag"
)

var DirectoryNotEmpty = errors.New("Directory Not Empty")
var ProtonNotFound = errors.New("Proton Not Found in Steam Libraries")

const colorRed = "\033[0;31m"
const colorGreen = "\033[0;32m"
const colorNone = "\033[0m"

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
	workingdir                   string //working dir of windows executable
	exeargs                      string //i.e.: -no-splash
}

type proton struct {
	id                string
	libpath           string
	name              string
	installfoldername string
}

func (p protonrunable) ConvertToCall() string {
	return fmt.Sprintf(" cd \"$(dirname \"$(realpath -- \"%s\")\")\" &&      STEAM_COMPAT_CLIENT_INSTALL_PATH=\"%s\" STEAM_COMPAT_DATA_PATH=\"%s\" \"%s/proton\" run \"%s\" %s", p.workingdir, p.steamcompatclientinstallpath, p.steamcompatdatapath, p.protonpath, p.exepath, p.exeargs)
}

func MakeProtonRunable(s steaminstance, compatpath string, protonu proton, exepath string, exeargs string) protonrunable {
	//envargs done
	steamcompatclientinstallpath := s.install
	steamcompatdatapath := compatpath
	protonpath := protonu.installfoldername
	//exepath done
	//exeargs done
	p := protonrunable{steamcompatclientinstallpath: steamcompatclientinstallpath, steamcompatdatapath: steamcompatdatapath, protonpath: protonpath, exepath: exepath, exeargs: exeargs, workingdir: filepath.Dir(exepath)}
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

func AskYesNo(prompt string) bool {
	fmt.Fprintf(os.Stdout, "%s [%sY%s/%sN%s]", prompt, colorGreen, colorNone, colorRed, colorNone)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, "\n")
	if text == "y" || text == "Y" {
		return true
	}
	return false
}

func MakePrefixIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil { // directory does not exist
		if AskYesNo("CompatData path not found, create path: \"" + path + "\"?") {
			os.MkdirAll(path, os.FileMode(0777))
		} else {
			panic(errors.New("raa"))
		}

	}
	dircontlist, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	if len(dircontlist) > 0 {
		good := false
		for _, item := range dircontlist {
			if item.Name() == "pfx.lock" {
				good = true
				break
			}
		}

		if good == false {
			return DirectoryNotEmpty
		}
	}
	return nil

}

func GetSteamGameList(s steaminstance) ([]game, []proton) {
	var gamelist []game
	var protonlist []proton
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	fpath := dirname + "/.steam/steam/steamapps/libraryfolders.vdf"
	fpath = s.install + "/steamapps/libraryfolders.vdf"
	//fmt.Println(fpath)
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
					//fmt.Println("Missing Library @ " + location + " @ Lookup of " + gameid)
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
					//fmt.Println("Missing Library @ " + location + " @ Lookup of " + gameid)
				}

			}

		}

	}

	return gamelist, protonlist
}

// func GetListOfProtons

func GetIndexOfSelectedProtonFromId(prlist []proton, id int) (int, error) {
	for idv, k := range prlist {
		if k.id == strconv.Itoa(id) {
			return idv, nil
		}

	}
	return -1, ProtonNotFound
}

func HelpText() {
	helptext := `Plauncher is a simple command line tool designed to make running non-Steam executables through Proton as simple as using Wine.
Options:
-h, --help: 		Disply this help text
-c, --compat	    Path to Proton Compatdata Folder
-p, --proton		SteamID of selected Proton (defaults to Proton Experimental)
 _, --steampath     Path to steam instance
-o, --printcommand  Print command instead of running it
`
	fmt.Println(helptext)

}

func main() {
	//setup default values

	uhomedir, _ := os.UserHomeDir()
	defaultprotoncompatdata := uhomedir + "/plauncherprefix"
	defaultsteaminstancepath := uhomedir + "/.steam/steam/"
	defaultprotonid := 1493710
	//setup flags
	var protoncompatdata = "neverrep"
	var steaminstancepath = "neverrep"
	var protonid = 0
	var helpmode bool = false
	var printcommand bool = false
	flag.StringVarP(&protoncompatdata, "compat", "c", defaultprotoncompatdata, "Path to Proton prefix")
	flag.StringVar(&steaminstancepath, "steampath", defaultsteaminstancepath, "Path to Steam instance")
	flag.IntVarP(&protonid, "proton", "p", defaultprotonid, "SteamID of selected Proton (Defualt is Proton Experimental)")
	flag.BoolVarP(&helpmode, "help", "h", false, "Display Help")
	flag.BoolVarP(&printcommand, "printcommand", "o", false, "Print the commnand instead of running it")
	//Get flags

	flag.Parse()
	remainingargs := flag.Args()
	if len(remainingargs) == 0 {
		helpmode = true
		fmt.Println("Error, no Windows executable supplied")
	}

	if helpmode {
		HelpText()
		return
	}
	err := MakePrefixIfNotExist(protoncompatdata)
	if err != nil {
		panic(err)
	}

	mainsteaminstance := steaminstance{install: steaminstancepath}
	a, p := GetSteamGameList(mainsteaminstance)

	protonindex, err := GetIndexOfSelectedProtonFromId(p, protonid)
	if err != nil {
		panic(err)
	}

	a = a // we dont want to get rid of this, otherwise a is unused (golang momment)

	//BetterPrintGameList(a)
	//BetterPrintProtonList(p)

	ex := MakeProtonRunable(mainsteaminstance, protoncompatdata, p[protonindex], remainingargs[0], "")
	call := ex.ConvertToCall()

	if printcommand {
		fmt.Println(ex)
		fmt.Println(call)
	} else {
		e := exec.Command("bash", "-c", call)
		var stdBuffer bytes.Buffer
		mw := io.MultiWriter(os.Stdout, &stdBuffer)
		e.Env = os.Environ()
		e.Stdout = mw
		e.Stderr = mw
		e.Run()
	}
}
