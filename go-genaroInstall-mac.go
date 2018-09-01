package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var exeFilename = "go-genaro"
var tmpZip = exeFilename+"Install-mac.zip"
var password []byte
var bootnodes string

var portDefault = "30315"
var wsportDefault = "8545"

type address struct {
	Address string `json:"address"`
}

var dir string
var keystore string

var port string
var wsport string


type parameter struct {
	Dir string `json:"dir"`
	Keystore string `json:"privatekey"`
	Password string `json:"Password"`
	Port 	string	`json:"port"`
	Wsport	string	`json:"wsport"`
}


func main() {
	fmt.Println("Please wait while installing")
	JsonParse := NewJsonStruct()
	configParameter := parameter{}
	JsonParse.Load("./config", &configParameter)
	if"" == configParameter.Dir || "" == configParameter.Keystore || "" == configParameter.Password {
		fmt.Println("config parameters cannot be empty")
		return
	}
	dir = configParameter.Dir
	if "" == dir {
		fmt.Println("Input directory failed")
		return
	}
	var chainNode = dir + "chainNode"
	var privateKey = dir + "chainNode/keystore/privateKey"


	keystore = configParameter.Keystore

	if "" == keystore {
		fmt.Println("Input private key address failed")
		return
	}
	password := configParameter.Password
	if "" == password {
		fmt.Println("please enter password:")
		return
	}

	wsport = configParameter.Wsport
	if "" == wsport {
		wsport = wsportDefault
	}
	port = configParameter.Port
	if "" == port {
		port = portDefault
	}


	exist, err := PathExists(dir)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return
	}

	if exist {
		err = os.RemoveAll(dir)
		if nil != err {
			fmt.Println("Remove genaroNetwork error!")
		}
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("mkdir error!")
		return
	}
	if ioutil.WriteFile(dir+"password", []byte(password), 0644) != nil {
		fmt.Println("Password write failed")
		return
	}




	cmd := exec.Command("cp", tmpZip, dir)
	_, err = cmd.CombinedOutput()

	if err != nil {
		fmt.Println("authorization failure")
		return
	}

	unzip()
	cmd = exec.Command("chmod", "777", dir+exeFilename)
	_, err = cmd.CombinedOutput()

	if err != nil {
		fmt.Println("authorization failure")
		return
	}

	cmd = exec.Command("chmod", "777", dir+exeFilename+"Restart-mac")
	_, err = cmd.CombinedOutput()

	if err != nil {
		fmt.Println("authorization failure")
		return
	}

	//time.Sleep(time.Duration(2)*time.Second)

	cmd = exec.Command(dir+exeFilename, "init", dir+"genaro.json", "--datadir", dir+"chainNode")
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("build chainNode failure")
		return
	}

	result := copyFile(keystore, privateKey)
	if false == result {
		fmt.Println(err)
		fmt.Println("copy keystore error")
		return
	}

	cmd = exec.Command("chmod", "0644", privateKey)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println("authorization failure")
		return
	}

	privateKeyRes, err := os.Open(privateKey)
	if err != nil {
		fmt.Println(err)
		fmt.Println("read privateKey error")
		return
	}
	defer privateKeyRes.Close()
	privateKeyResult, err := ioutil.ReadAll(privateKeyRes)
	if nil != err {
		fmt.Println("read privateKey error")
	}

	var accountAdderss address
	err = json.Unmarshal([]byte(string(privateKeyResult)), &accountAdderss)
	if nil != err {
		fmt.Println("get account Adderss error")
		return
	}

	if "" == accountAdderss.Address {
		fmt.Println("get account Adderss error")
		return
	}

	fileObj, err := os.Open(dir + "bootnodes")

	if err != nil {
		fmt.Println("open bootnodes error")
	}
	defer fileObj.Close()
	if contents, err := ioutil.ReadAll(fileObj); err == nil {
		bootnodes = strings.Replace(string(contents), "\n", "", 1)
	} else {
		fmt.Println("get bootnodes error")
	}

	parameter := startupParameter{
		Dir:dir,
		ChainNode:chainNode,
		Address:accountAdderss.Address,
		Bootnodes:bootnodes,
		Port:port,
		Wsport:wsport,
	}

	parameterJson,err := json.Marshal(parameter)
	if nil != err {
		fmt.Println("json.Marshal(parameter) error")
		return
	}
	if ioutil.WriteFile(dir+"startupParameter",parameterJson, 0644) != nil {
		fmt.Println("parameterJson write failed")
		return
	}
	fmt.Println("bootnodes: ",bootnodes)
	cmd = exec.Command(dir+exeFilename, "--ws", "--wsorigins=*", "--wsapi", "net,admin,personal,miner", "--datadir", chainNode, "--port", port, "--wsport", wsport, "--wsaddr", "127.0.0.1", "--unlock", "0x"+accountAdderss.Address, "--password", dir+"password", "--syncmode", "full", "--mine", "--bootnodes", bootnodes)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Installed and started successfully")

}

type startupParameter struct {
	Dir string `json:"dir"`
	ChainNode string `json:"chainNode"`
	Address string `json:"address"`
	Bootnodes string `json:"bootnodes"`
	Port 	string	`json:"port"`
	Wsport	string	`json:"wsport"`
}


func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func unzip() bool {
	fmt.Println(dir + tmpZip)
	r, err := zip.OpenReader(dir + tmpZip)
	if err != nil {
		fmt.Println(err)
		return false
	}
	for _, k := range r.Reader.File {
		if strings.HasPrefix(k.Name, "__MACOSX") {
			continue
		}
		if k.FileInfo().IsDir() {
			err := os.MkdirAll(k.Name, 0644)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
		r, err := k.Open()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("unzip: ", k.Name)
		defer r.Close()
		NewFile, err := os.Create(dir + k.Name)
		if err != nil {
			fmt.Println(err)
			continue
		}
		io.Copy(NewFile, r)
		NewFile.Close()
	}
	return true
}

func copyFile(source, dest string) bool {
	if source == "" || dest == "" {
		fmt.Println("source or dest is null")
		return false
	}
	source_open, err := os.Open(source)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer source_open.Close()
	dest_open, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 644)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer dest_open.Close()
	_, copy_err := io.Copy(dest_open, source_open)
	if copy_err != nil {
		fmt.Println(copy_err.Error())
		return false
	} else {
		return true
	}
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