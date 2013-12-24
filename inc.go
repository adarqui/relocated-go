package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	STATUS_INPROGRESS = 1
	STATUS_FINISHED   = 2
	STATUS_COMPLETED  = 3
)

type JsonType struct {
	Root ObjectType
}

type ObjectType struct {
	Daemon     bool
	MaxProc    int
	/*
	 * IntervalPoll = seconds to poll, ie, poll every 1 second for changes
	 * IntervalElapsed = seconds elapsed which trigger a relocation
	 */
	IntervalPoll   time.Duration
	IntervalElapsed   time.Duration
	Relocators []RelocatorsType
}

type RelocatorsType struct {
	Name        string
	NameSpace   string
	Class       string
	Destination string
	Glob        []string
	Relocate    string
	Interval    time.Duration
	Files       map[string]RelocatorsFileType
}

type RelocatorsFileType struct {
	Path      string
	Status    int
	Relocator RelocatorsType
	Fi_now    os.FileInfo
	Fi_prev   os.FileInfo
	Last      time.Time
}

func (x *JsonType) Sanitize() {
	if x.Root.MaxProc <= 0 {
		x.Root.MaxProc = 10
	}
	if x.Root.IntervalPoll <= 0 {
		x.Root.IntervalPoll = 5
	}
	if x.Root.IntervalElapsed <= 0 {
		x.Root.IntervalElapsed = 60
	}

	x.Root.IntervalPoll = x.Root.IntervalPoll * time.Second;
	x.Root.IntervalElapsed = x.Root.IntervalElapsed * time.Second

	log.Printf("Global config settings: MaxProc=%i, IntervalPoll=%i, IntervalElapsed=%i\n", x.Root.MaxProc, x.Root.IntervalPoll, x.Root.IntervalElapsed)

	for key, element := range x.Root.Relocators {
		if element.Interval <= 0 {
			element.Interval = x.Root.IntervalElapsed
		} else {
			element.Interval = element.Interval * time.Second
		}
		x.Root.Relocators[key] = element

		log.Printf("Relocator: Name=%s, NameSpace=%s, Class=%s, Destination=%s, Interval=%i, Glob=%v\n", element.Name, element.NameSpace, element.Class, element.Destination, element.Interval, element.Glob)

	}
}

func (x *JsonType) Init() {

	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Init:ReadFile", err)
	}

	err = json.Unmarshal(file, x)
	if err != nil {
		log.Fatal("Init:Unmarshal:Json is invalid", err)
	}

	exec := make(chan RelocatorsFileType, 32)

	x.Sanitize()

	if x.Root.Daemon == true {
		Daemon()
	}

	for _, element := range x.Root.Relocators {
		element_copy := element
		go Init_Watcher(x.Root.IntervalPoll, &element_copy, exec)
	}

	for i := 0; i < x.Root.MaxProc; i++ {
		go x.Init_ExecPool(exec, i)
	}

	select {}
}

func Init_Watcher(Poll time.Duration, x *RelocatorsType, c chan RelocatorsFileType) {

	x.Files = map[string]RelocatorsFileType{}
	for {
		x.Fill_Glob(c)
		time.Sleep(Poll)
	}
}

func (x *RelocatorsType) Fill_Glob(c chan RelocatorsFileType) {

	for key, element := range x.Files {
		/*
		 * Clean up any members of the map whose state is FINISHED or COMPLETED
		 */
		if element.Status == STATUS_FINISHED || element.Status == STATUS_COMPLETED {
			err := Stat_File(&element)
			if err != nil {
				delete(x.Files, key)
			}
		}
	}

	for _, element := range x.Glob {

		/*
		 * Iterate over all of the Glob paths & the results from globbing
		 */

		matches, err := filepath.Glob(element)
		if err != nil {
			/* No matches */
			continue;
		}

		for _, element := range matches {
			/*
			 * Matches found, look them up (pre-existing) or create new members of the Files map: Stat() them to keep track of file size changes.
			 */
			if _, ok := x.Files[element]; !ok {
				/* New entry, not found in Files map */
				var F RelocatorsFileType
				F.Status = STATUS_INPROGRESS
				F.Relocator.Clone_Relocator(x)
				F.Path = element
				F.Last = time.Now()
				err = Stat_File(&F)
				if err != nil {
					continue
				}
				x.Files[element] = F
			} else {
				/* Entry already exists, update it & decide whether or not we should relocate it */
				F := x.Files[element]

				err = Stat_File(&F)
				if err != nil {
					delete(x.Files, element)
					continue
				}

				if F.Status == STATUS_FINISHED {
					/* Already triggered for relocation, but it might still be moving (copying over a two mountpoints etc) */
					continue
				}

				now := time.Now()
				if F.Fi_now.Size() == F.Fi_prev.Size() && now.Sub(F.Last) > x.Interval  {
					/* Interval has passed & the current/previous size match. This triggers relocation */
					F.Status = STATUS_FINISHED
					x.Files[element] = F
					c <- F
				}
			}

		}
	}
}

func (x *RelocatorsType) Clone_Relocator(y *RelocatorsType) {
	x.Name = y.Name
	x.NameSpace = y.NameSpace
	x.Class = y.Class
	x.Destination = y.Destination
	x.Relocate = y.Relocate
	return
}

func Stat_File(x *RelocatorsFileType) error {
	x.Fi_prev = x.Fi_now
	var err error
	x.Fi_now, err = os.Stat(x.Path)
	return err
}

func (x *JsonType) Init_ExecPool(c chan RelocatorsFileType, i int) {

	for message := range c {
		/* We have received a message to relocate the file, as part of our ExecPool. So, execute the custom relocation script. This script should process & then move the file to it's destination. Moving it clears the file out of the Files map */
		cmd := exec.Command(message.Relocator.Relocate, message.Relocator.Name, message.Relocator.NameSpace, message.Relocator.Class, message.Path, message.Relocator.Destination)
		log.Printf("ExecPool: relocating %s for %s:%s:%s to %s\n",
			message.Path, message.Relocator.Name, message.Relocator.NameSpace, message.Relocator.Class, message.Relocator.Destination)

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Init_ExecPool:cmd.Run():%v\n", err)
		}
	}
}
