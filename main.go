package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ip, port, configFile, outputFile string
	verboseFlag                      bool

	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger

	reqRespMap map[string]HeaderAndBody
)

type HeaderAndBody struct {
	Header string
	Body   string
}

func readConfiguration() bool {

	reqRespMap = make(map[string]HeaderAndBody)

	buf, err := ioutil.ReadFile(configFile)

	if err != nil {
		fmt.Printf("Error occured while reading file %v: %v", configFile, err)
		return false
	}

	err = yaml.Unmarshal(buf, reqRespMap)

	if err != nil {
		fmt.Printf("in file %q: %v", configFile, err)
		return false
	}

	return true
}

func initializeAndStartListening() {

	// Initiate Logger
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	readConfiguration()

	if verboseFlag {
		fmt.Printf("Configuration:\n%v\n", reqRespMap)
	}

	http.HandleFunc("/", defaultHandler)
	fmt.Println("listing on " + ip + ":" + port + " ...")
	http.ListenAndServe(ip+":"+port, nil)
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {

	key := req.Method + "|" + strings.Trim(req.RequestURI, "/")
	response := reqRespMap[key]

	if response == (HeaderAndBody{}) {
		response = reqRespMap["default|default"]
	}

	// -- Write HTTP Response Code -- //
	w.WriteHeader(http.StatusOK)

	// -- Process Response Header -- //

	responseText, err := ioutil.ReadFile(response.Header)

	if err != nil {
		fmt.Printf("Error occured while sending response header .. %v: %v\n", configFile, err)
		ErrorLogger.Printf("Error occured while sending response header .. %v\n", err)
	} else if strings.Trim(string(responseText), " ") != "" {

		headers := strings.Split(string(responseText), "\n")

		for _, line := range headers {

			keyValue := strings.Split(line, ":")
			_key := strings.Trim(keyValue[0], " ")
			_value := strings.Trim(keyValue[1], " ")
			w.Header().Add(_key, _value)
			fmt.Printf("{key: %v value: %v}\n", _key, _value)
		}
	}

	fmt.Printf("Printing w.Header(): %v\n", w.Header())

	// -- Process Response Body -- //

	responseText, err = ioutil.ReadFile(response.Body)

	if err != nil {
		fmt.Printf("Error occured while sending response body .. %v: %v\n", configFile, err)
		ErrorLogger.Printf("Error occured while sending response body .. \n%v", err)
	}

	w.Write(responseText)

	//------- logging code ------//

	if verboseFlag {
		fmt.Println("---- Serving Incoming request ---- ")
		for name, headers := range req.Header {
			for _, h := range headers {
				if verboseFlag {
					fmt.Printf("%v: %v\n", name, h)
				}
			}
		}
		fmt.Printf("HTTP Header to be served: %v\n", response.Header)
		fmt.Printf("HTTP Body to be served: %v\n", response.Body)
		fmt.Println("---------------End---------------- ")
	}

}

func main() {

	if !readArguments() {
		fmt.Printf("Usage:\n\thttp-mock-server -ip <IPAddress> -port <PORT> -config <config-file-name> [-output <output-file-name>] [-verbose <true|false>]\n")
		os.Exit(-1)
	}
	initializeAndStartListening()

}

func readArguments() bool {

	args := os.Args[1:]
	argsLength := len(args)

	_ip := flag.String("ip", "", "IP Address where applicaiton will bind, if none provided, it binds to all interface as default")
	_port := flag.String("port", "8080", "TCP port wher eapplication will listen, default port is 8080")
	_configFile := flag.String("config", "", "Mandatory input parameter which contains path, request and response to serve")
	_outputFile := flag.String("output", "output.log", "output file where application log will be written. by default it will write in default directory")
	_verboseFlag := flag.Bool("verbose", false, "by default verbose will be truned off")

	flag.Parse()

	ip = *_ip
	port = *_port
	configFile = *_configFile
	outputFile = *_outputFile
	verboseFlag = *_verboseFlag

	if ip == "" {
		ip = "localhost"
	}

	if verboseFlag {
		fmt.Printf("ip : %v\nport : %v\nconfigFile : %v\nOoutputFile : %v\nverboseFlag : %v\n", ip, port, configFile, outputFile, verboseFlag)
	}

	return argsLength > 0
}
