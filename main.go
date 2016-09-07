package main

import (
	"fmt"
	"flag"
	"strings"
	"os"
	"bufio"
	"io"
	"encoding/json"
	"encoding/hex"
)

var filename string

func init() {
	flag.StringVar(&filename, "FileName", "DiscoveryAndLoginCommand", "Source file. 'DiscoveryAndLoginCommand' by default.")
}

type outputStruct struct {
	Name	string
	Data	[]byte
}

func (l *outputStruct)fillName(s []byte)  {
	l.Name = strings.TrimSpace(string(s[90:]))
	return
}

func (l *outputStruct)addLine(s []byte)  {
	i := 0
	for _, el := range s {
		if el != 0x20 {
			s[i] = el
			i++
		}
	}
	m := make([]byte, len(s[:i])/2)
	_, _ = hex.Decode(m, s[:i])

	l.Data = append(l.Data, m...)
	return
}

func (l outputStruct)String()(string) {
	return fmt.Sprintf("Name=%s\nData=% x", l.Name, l.Data)
}
func main() {

	flag.Parse()
	if filename == "" {
		fmt.Println("File name is empty.")
		return
	}

	fiRead, err := os.Open(filename)
	defer fiRead.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fiWrite, err := os.Create(filename+ ".json")
	defer fiWrite.Close()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("File", filename+ ".json", "created.")
	}
	fileRead := bufio.NewReader(fiRead)
	fileWrite := bufio.NewWriter(fiWrite)
	defer fileWrite.Flush()
	encJSON := json.NewEncoder(fileWrite)

	var outElement outputStruct
	flagName := false
	flagData := false

	for {
		sliceBlock, _, errR := fileRead.ReadLine()
		if errR == io.EOF {
			return
		}
		if errR != nil && errR != io.EOF {
			fmt.Println(errR)
			return
		}

		if len(sliceBlock) < 6 {
			flagData = true
			continue
		}
		if string(sliceBlock[:5]) == "Frame" {
			flagData = false
		}
		if strings.TrimSpace(string(sliceBlock[:5])) == "No." {
			flagName = true
			continue
		}
		if flagName {
			outElement = outputStruct{Name:"", Data:[]byte{}}
			outElement.fillName(sliceBlock)
			flagName = false
		}
		if string(sliceBlock[:11]) == "Reassembled" {
			flagData = true
		}
		if string(sliceBlock[:6]) == "0000  " && flagData {
			for string(sliceBlock[:6]) != "No.   " && len(sliceBlock) > 6 {
				outElement.addLine(sliceBlock[6:54])
				sliceBlock, _, errR = fileRead.ReadLine()
				if errR == io.EOF {
					break
				}
				if errR != nil && errR != io.EOF {
					fmt.Println(errR)
					return
				}
			}
			flagData = false

			if err := encJSON.Encode(outElement); err != nil {
				fmt.Println(err)
			}

		}
	}

	return
}
