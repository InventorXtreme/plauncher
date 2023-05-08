package main

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func hehe(prog int) {
	fmt.Print(prog)
}

type service struct {
	unit        string
	load        string
	active      string
	sub         string
	description string
}

func GetSysList() string {
	out, err := exec.Command("systemctl", "--type", "service", "list-units").Output()
	if err != nil {
		log.Fatal(err)
		fmt.Println("erred")
	}
	return string(out)
}

func pr(item *core.QModelIndex) {
	fmt.Println(item.Row())
}

func addservicetolist(svc service, ls *widgets.QListWidget) {
	ls.AddItem(svc.unit)
}

func addServiceListToListWidget(slist []service, ls *widgets.QListWidget) {
	for _, v := range slist {
		addservicetolist(v, ls)
	}

}

func main() {
	funnyargs := os.Args
	rargs := append(funnyargs, "--no-sandbox")
	// needs to be called once before you can start using the QWidgets
	app := widgets.NewQApplication(len(rargs), rargs)

	// create a windowm
	// with a minimum size of 250*200
	// and sets the title to "Hello Widgets Example"
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(250, 200)

	currentlist := widgets.NewQListWidget(nil)

	currentlist.ConnectDoubleClicked(pr)

	servstring := GetSysList()
	servllist := strings.Split(servstring, "\n")

	descpointgetter := regexp.MustCompile(" D")
	loadpointgetter := regexp.MustCompile(" LO")
	activepointgetter := regexp.MustCompile(" AC")
	subpointgetter := regexp.MustCompile(" SUB")

	descriptionpoint := descpointgetter.FindStringIndex(servllist[0])[0] + 1
	loadpoint := loadpointgetter.FindStringIndex(servllist[0])[0] + 1
	activepoint := activepointgetter.FindStringIndex(servllist[0])[0] + 1
	subpoint := subpointgetter.FindStringIndex(servllist[0])[0] + 1

	var unitlist []service

	for _, s := range servllist {
		if len(s) > descriptionpoint {
			unitlist = append(unitlist, service{s[1:loadpoint], s[loadpoint:activepoint], s[activepoint:subpoint], s[subpoint:descriptionpoint], s[descriptionpoint:]})
		}
	}

	fmt.Println(unitlist[2].unit)

	addServiceListToListWidget(unitlist, currentlist)

	window.SetCentralWidget(currentlist)
	// make the window visible
	window.Show()

	// start the main Qt event loop
	// and block until app.Exit() is called
	// or the window is closed by the user
	app.Exec()
}
