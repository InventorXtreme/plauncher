package main

import "fmt"

func BetterPrintGameList(g []game) {
	for _, k := range g {
		fmt.Printf("GAMEID: %s\n", k.id)
		fmt.Printf("LIBPATH: %s\n", k.libpath)
		fmt.Printf("NAME: %s\n", k.name)
		fmt.Printf("COMPATPATH: %s\n", k.compatpath)
		fmt.Printf("GAME INSTALL DIR: %s\n", k.installfoldername)
		fmt.Println("")

	}
}

func BetterPrintProtonList(g []proton) {
	for _, k := range g {
		fmt.Printf("GAMEID: %s\n", k.id)
		fmt.Printf("LIBPATH: %s\n", k.libpath)
		fmt.Printf("NAME: %s\n", k.name)
		fmt.Printf("PROTON INSTALL DIR: %s\n", k.installfoldername)
		fmt.Println("")

	}
}
