package main

import  (
	"fmt"
	"os"
	"github.com/andygrunwald/vdf"
	"errors"
)

type game struct {
	id string
	libpath string
	name string
}

type steaminstance struct {
	install string
	proton string
}

func GetMapFrom(source map[string]interface{},key string) map[string]interface{}  {
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

func GetNameFromId(id string,path string) ( string ,error) {
	filepath :=  path+"/steamapps/appmanifest_"+id + ".acf"

	f, err := os.Open(filepath)
	if err != nil {
		return "", errors.New("FileRead Err")
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m,err := p.Parse()
	if err != nil {
		panic(err)

	}

	name, err := GetValFrom(GetMapFrom(m,"AppState"),"name")
	if err != nil {
		panic(err)
	}
	return name , nil


}

func IsGameWindows(id string,path string) ( bool ,error) {
	filepath :=  path+"/steamapps/appmanifest_"+id + ".acf"

	f, err := os.Open(filepath)
	if err != nil {
		return false, errors.New("FileRead Err")
	}
	defer f.Close()

	p := vdf.NewParser(f)
	m,err := p.Parse()
	if err != nil {
		panic(err)

	}
	platform, _ := GetValFrom(GetMapFrom(GetMapFrom(m,"AppState"),"UserConfig"),"platform_override_source")
	if platform == "windows" {
		return true, nil
	} else {
		return false, nil
	}



}

// func GetCompatDataPathForGame(g game, i steaminstance) string {

// }

func (g game) GetRunCommand(steam steaminstance, exefile string ) string {
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
	set := GetMapFrom(m,"libraryfolders")

	for k,_ := range set {
		currentlibrary := GetMapFrom(set,k)
		location, err := GetValFrom(currentlibrary,"path")
		if err != nil {
			panic(err)
		}
		fmt.Println(location)
		
		currentlibrary_applist := GetMapFrom(currentlibrary,"apps")


		for g,_ := range currentlibrary_applist {
			gameid := g
			gamelocation := location
			gamename, err := GetNameFromId(gameid,gamelocation)

			if err == nil {
				gamelist = append(gamelist,game{gameid,gamelocation,gamename})
				e, _ := IsGameWindows(gameid,gamelocation)
				if e == true {

					fmt.Println("Win: " + gameid + " " + gamelocation + " " + gamename)
				}
			}


		}

		

	}

	return gamelist
}

func main(){
	GetSteamGameList()
}
