package main

import (
	"fmt"
	"gopkg.in/qml.v1"
	"os"
	"runtime"
	"regexp"
	"bufio"
	"strings"
	"strconv"
	"io"
	"os/exec"
	"sync"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"flag"
	"path/filepath"
)

//Error handler, all bad errors will be sent here.
func handle(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	runtime.LockOSThread()
	if err := qml.Run(run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type Card struct {
	Name string
	Amount int
	Image string
	Loaded bool
}

type Deck struct {
	sync.Mutex
	list []Card
	game string
	filename string
	description string
	gamestring string
}

func (cards *Deck) Name(index int) string {
	cards.Lock()
	defer cards.Unlock()
	
	if index < 0 {
		return "Nil"
	}
	if index > len(cards.list)-1 {
		return "Nil"
	}
	return cards.list[index].Name
}

func (cards *Deck) Total() string {
	cards.Lock()
	defer cards.Unlock()
	
	total := 0
	
	for _, v := range cards.list {
		total += v.Amount
	}
	
	return fmt.Sprint(total)
}


func (cards *Deck) Setname(index int, name string) {
	cards.Lock()
	defer cards.Unlock()
	
	if index < 0 {
		return
	}
	if index > len(cards.list)-1 {
		return
	}
	
	cards.list[index].Name = name 
}

func (cards *Deck) Setamount(index int, amount string) {
	cards.Lock()
	defer cards.Unlock()
	
	if index < 0 {
		return
	}
	if index > len(cards.list)-1 {
		return
	}
	
	i, err :=  strconv.ParseInt(amount, 10, 0)
	handle(err)
	
	cards.list[index].Amount = int(i)
}

func (cards *Deck) Amount(index int) string {
	cards.Lock()
	defer cards.Unlock()
	
	if index < 0 {
		return "Nil"
	}
	if index > len(cards.list)-1 {
		return "Nil"
	}
	return fmt.Sprint(cards.list[index].Amount)
}

func (cards *Deck) Len() int {
	cards.Lock()
	defer cards.Unlock()
	
	return len(cards.list)
}

func (cards *Deck) Game() string {
	cards.Lock()
	defer cards.Unlock()
	
	return cards.game
}
func (cards *Deck) Loaded(index int) bool {
	cards.Lock()
	defer cards.Unlock()
	
	
	if index < 0 {
		return false
	}
	if index > len(cards.list)-1 {
		return false
	}
	return cards.list[index].Loaded
}

func (cards *Deck) Image(index int) string {
	cards.Lock()
	defer cards.Unlock()
	if index < 0 {
		return ""
	}
	if index > len(cards.list)-1 {
		return "Nil"
	}
	return cards.list[index].Image
}

func (cards *Deck) Add() {
	cards.Lock()
	defer cards.Unlock()

	cards.list = append(cards.list, Card{Amount:1})
}


func (cards *Deck) Remove(i int) {
	cards.Lock()
	defer cards.Unlock()

	cards.list = append(cards.list[:i], cards.list[i+1:]...)

}

var lastLoaded int = -1

func (cards *Deck) Load(index int) {
	go func(index int) {
		if index < 0 {
			return
		}
		if index > len(cards.list)-1 {
			return 
		}
		
		if index == lastLoaded {
			return
		} 
		
		if strings.TrimSpace(cards.Name(index)) == "" {
			cards.list[index].Image = ""
			cards.list[index].Loaded = true
			return
		}
		
		if _, err := os.Stat(filepath.Dir(cards.filename) + "/cards/" + cards.game + "/" + cards.Name(index) + ".jpg"); !os.IsNotExist(err) {
			cards.list[index].Image = filepath.Dir(cards.filename) + "/cards/" + cards.game + "/" + cards.Name(index) + ".jpg"
			cards.list[index].Loaded = true
			return
		}
			
		command := "decker"
		if runtime.GOOS == "windows" {
			command, err = filepath.Abs(os.Args[0])
			command = filepath.Dir(command)
			if err != nil {
				command = "decker"
			} else {
				command += "/decker"
			}
		}			
	
		decker := exec.Command(command, "-d", cards.Name(index), "-g", cards.Game())
		data, err := decker.Output()
		if err != nil {
			fmt.Println(string(data), err.Error())
		}
		
		lines := strings.Split(string(data), "\n")
		
		Cards.Lock()
		defer Cards.Unlock()
		if len(lines)-2 > -1 {
			cards.list[index].Image = strings.TrimSpace(lines[len(lines)-2])
			cards.list[index].Loaded = true
		}
		lastLoaded = index
		
	}(index)
} 

func (cards *Deck) Save() {
	if file, err := os.Create(cards.filename); err == nil {
		
		file.WriteString(cards.gamestring+"\n\n")
		file.WriteString(cards.description+"\n\n")
		for _, v := range cards.list {
			file.WriteString(fmt.Sprint(v.Amount)+" "+v.Name+"\n")
		}
		file.Close()
	} else {
		fmt.Println(err.Error())
	}
}

func (deck *Deck) Open(filename string) {
	//get rid of the "file://"
	
	
	filename = filename[7:]
	lastLoaded = -1

	go func(filename string) {
		command := "decker"
		if runtime.GOOS == "windows" {
			command, err = filepath.Abs(os.Args[0])
			command = filepath.Dir(command)
			if err != nil {
				command = "decker"
			} else {
				command += "/decker"
			}
		}
	
		decker := exec.Command(command, "-I", filename)
		data, err := decker.Output()
		handle(err)
		Cards.Lock()
		defer Cards.Unlock()
		Cards.game = strings.TrimSpace(string(data))
	}(filename)

	cards := new(Deck)
	cards.list = make([]Card, 0)

	//Compile a regular expression to test if the line is a card
	r, _ := regexp.Compile("^((\\d+x)|(x?\\d+)) +[^ \n]+")
	
	if file, err := os.Open(filename); err == nil {

		//Read the first line and trim the space.
		reader := bufio.NewReader(file)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		
		cards.gamestring = line
	
		//Loop through the file.
		for {
			line, err := reader.ReadString('\n') //Parse line-by-line.
			if err == io.EOF {
				break
			}
			handle(err)

			//Trim the spacing. TODO trim spacing in between words that are used for nice reading.
			line = strings.TrimSpace(line)


			//Check if the line is a card 
			// (nx, n or xn followed by at least one space and then anything not space)
			if r.MatchString(line) {
				//We need to seperate the name from the number of cards.
				//This does that.
				r, _ := regexp.Compile("^((\\d+x)|(x?\\d+))")

				name := r.ReplaceAllString(line, "");
				name = strings.Join(strings.Fields(name), " ")
		
				count, _ := strconv.Atoi(strings.Replace(r.FindString(line), "x", "", -1));
		
				cards.list = append(cards.list, Card{Name:name,Amount:count})
			} else {
				if strings.TrimSpace(line) != "" {
					cards.description += line+"\n"
				}
			}
		}
		file.Close()
	} else {
		fmt.Println(err)
	}
	Cards.list = cards.list
	Cards.filename = filename
	Cards.description = cards.description
	Cards.gamestring = cards.gamestring
}

var Cards *Deck = new(Deck)

func run() error {
	flag.Parse()

	engine := qml.NewEngine()
	
	engine.AddImageProvider("card", func(id string, width, height int) image.Image {
		if id == "" {
			return  image.NewNRGBA(image.Rect(0, 0, 1, 1))
		}
	
		f, err := os.Open(id)
		if err != nil {
			fmt.Println("error: ", err.Error())
			return image.NewNRGBA(image.Rect(0, 0, 0, 0))
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			fmt.Println("error: ", err.Error())
			
			return  image.NewNRGBA(image.Rect(0, 0, 0, 0))
		}
		return img
	})

	
	context := engine.Context()
	
	Cards.list = make([]Card, 0)

	Cards.Open("file://"+flag.Arg(0))
	
	context.SetVar("cards", Cards)
	

	controls, err := engine.LoadFile("decker.qml")
	if err != nil {
		return err
	}

	window := controls.CreateWindow(nil)

	window.Show()
	
	window.Wait()
	return nil
}
