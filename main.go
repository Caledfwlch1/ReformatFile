package main

import (
	"fmt"
	"flag"
	"runtime"
	"strings"
	"os"
	"bufio"
	"io"
	"bytes"
	"encoding/hex"
	"encoding/json"
)

var filename string

func init() {
	flag.StringVar(&filename, "FileName", "DiscoveryAndLoginCommand", "Source file. 'DiscoveryAndLoginCommand' by default.")
}

type outputStruct struct {
	Name	string
	Data	[]byte
}

func (l *outputStruct)fillData(s [][]byte)  {
	l.Name = string(s[1][115:])
	for _, i := range s[3:] {
		if len(i) < 54 {
			continue
		}
		l.Data = append(l.Data, stringToByte(i[6:54])...)
	}
	return
}

func (l outputStruct)String()(string) {
	return fmt.Sprintf("Name=%s\nData=% x", l.Name, l.Data)
}

func stringToByte(in []byte)([]byte) {
	in = bytes.TrimSpace(in)
	in = append(in, 0x32)
	j, i := 0, 0
	s := make([]byte, 2)
	out := make([]byte, len(in)/3)

	for {
		_, _ = hex.Decode(s, in[i:i+2])
		out[j] = s[0]
		j += 1
		i += 3
		if i >= len(in) {
			break
		}
	}
	return out
}

func main() {
	fmt.Println("Start.")
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

	fiWrite, err := os.Create(filename+ "_new")
	defer fiWrite.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileRead := bufio.NewReader(fiRead)
	fileWrite := bufio.NewWriter(fiWrite)
	encJSON := json.NewEncoder(fileWrite)
	for {
		block, errR := fileRead.ReadString(0x0c)
		if errR != io.EOF && errR != nil {
			fmt.Println(errR)
			return
		}

		indexReass := strings.Index(block, "Reassembled")
		indexFrame := strings.Index(block, "Frame")
		if indexReass < 1 {
			indexFrame = 2
			indexReass = indexFrame - 1
		}

		clearBlock := block[:indexFrame - 1] + block[indexReass:]
		sliceBlock := bytes.Split([]byte(clearBlock), []byte{0x0a})

		var outElement outputStruct

		outElement.fillData(sliceBlock)

		_ = encJSON.Encode(outElement)

		if errR == io.EOF {
			break
		}
	}
	fileWrite.Flush()


	return
}







// the function for debugging,
// it print function name, number of string and specified of variables
func PrintDeb(s ...interface{}) {
	name, line := procName(false, 2)
	fmt.Print("=> ", name, " ", line, ": ")
	fmt.Println(s...)
	return
}

// the function return the name of working function
func procName(shortName bool, level int) (name string, line int) {
	pc, _, line, _ := runtime.Caller(level)
	name = runtime.FuncForPC(pc).Name()
	if shortName {
		name = name[strings.Index(name, ".")+1:]
	}
	return name, line
}
