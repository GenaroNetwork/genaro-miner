package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os"
)

var exeFilename = "go-genaro"

func main() {
	JsonParse := NewJsonStruct()
	parameter := startupParameter{}
	JsonParse.Load("./startupParameter", &parameter)
	if "" == parameter.Address || "" == parameter.Dir || "" == parameter.ChainNode || "" == parameter.Bootnodes  || "" == parameter.Port || "" == parameter.Wsport{
		fmt.Println("Startup parameters cannot be empty")
		return
	}
	cmd := exec.Command(parameter.Dir+exeFilename, "--ws", "--wsorigins=*", "--wsapi", "net,admin,personal,miner", "--datadir", parameter.ChainNode, "--port", parameter.Port, "--wsport", parameter.Wsport, "--wsaddr", "127.0.0.1", "--unlock", "0x"+parameter.Address, "--password", parameter.Dir+"password", "--syncmode", "full", "--mine", "--bootnodes", parameter.Bootnodes)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("successful startup")
}

type startupParameter struct {
	Dir       string `json:"dir"`
	ChainNode string `json:"chainNode"`
	Address   string `json:"address"`
	Bootnodes string `json:"bootnodes"`
	Port 	string	`json:"port"`
	Wsport	string	`json:"wsport"`
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

type JsonStruct struct {
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}