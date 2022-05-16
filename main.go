package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	HttpCode int
	Delay    int
	Header   string
	Body     string
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
		fmt.Printf("Error occured while parsingin file %q: %v", configFile, err)
		ErrorLogger.Printf("Error occured while parsingin file %q: %v", configFile, err)
		return false
	}

	return true
}

func initializeAndStartListening() {

	fmt.Printf("Reading Configuration ...\n")
	InfoLogger.Printf("Reading Configuration ...\n")

	readConfiguration()

	http.HandleFunc("/", defaultHandler)

	fmt.Println("listing on " + ip + ":" + port + " ...")
	InfoLogger.Println("listing on " + ip + ":" + port + " ...")

	http.ListenAndServe(ip+":"+port, nil)
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {

	var logBuff strings.Builder

	fmt.Printf("Received request from ... %v\n", req.RemoteAddr)
	InfoLogger.Printf("Received request from ... %v\n", req.RemoteAddr)

	fmt.Printf("Dumping rquest data in output file ...\n")
	//------- logging code ------//

	fmt.Fprintf(&logBuff, "Request dump ..\n\n---- Incoming request trace ----\n\n")

	fmt.Fprintf(&logBuff, "Request URL:: %v\n", req.RequestURI)
	fmt.Fprintf(&logBuff, "HTTP Method:: %v\n", req.Method)

	fmt.Fprintf(&logBuff, "\nHeader:: \n")

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(&logBuff, "\t%v: %v\n", name, h)
		}
	}

	req.ParseForm()

	fmt.Fprintf(&logBuff, "\nrequest.Form::\n")

	for key, value := range req.Form {
		fmt.Fprintf(&logBuff, "\t%s : %s\n", key, value)
	}

	fmt.Fprintf(&logBuff, "\nrequest.PostForm::\n")
	for key, value := range req.PostForm {
		fmt.Fprintf(&logBuff, "\t%s : %s\n", key, value)
	}

	bodyStream, err := ioutil.ReadAll(req.Body)

	if err == nil {
		fmt.Fprintf(&logBuff, "\nrequest.Body::\n")
		fmt.Fprintf(&logBuff, "\n%v\n", string(bodyStream))

	} else {
		fmt.Printf("Error occured while reading request body .. %v\n", err)
		ErrorLogger.Printf("Error occured while reading request body .. \n%v", err)
	}

	fmt.Fprintf(&logBuff, "---------------End----------------\n")
	InfoLogger.Println(logBuff.String())

	if verboseFlag {
		fmt.Println(logBuff.String())
	}

	key := req.Method + "|" + strings.Trim(req.RequestURI, "/")
	response := reqRespMap[key]

	if response == (HeaderAndBody{}) {
		response = reqRespMap["default|default"]
	}

	fmt.Printf("Processing Request ...\n")
	InfoLogger.Printf("Processing Request ...\n")

	fmt.Printf("Sleeping for %v seconds ...\n", response.Delay)
	InfoLogger.Printf("Sleeping for %v seconds ...\n", response.Delay)
	time.Sleep(time.Duration(response.Delay) * time.Second)

	// -- Process Response Header -- //

	fmt.Printf("Writing Headers ... \n")
	InfoLogger.Printf("Writing Headers ... \n")

	responseText, err := ioutil.ReadFile(response.Header)

	if err != nil {
		fmt.Printf("Error occured while sending response header .. : %v\n", err)
		ErrorLogger.Printf("Error occured while sending response header .. %v\n", err)
	} else if strings.Trim(string(responseText), " ") != "" {

		headers := strings.Split(string(responseText), "\n")

		for _, line := range headers {

			keyValue := strings.Split(line, ":")
			_key := strings.Trim(keyValue[0], " ")
			_value := strings.Trim(keyValue[1], " ")
			w.Header().Add(_key, _value)
		}

		fmt.Printf("Writing Response Headers ... \n%v\n", string(responseText))
		InfoLogger.Printf("Writing Response Headers ... \n%v\n", string(responseText))

	}

	// -- Write HTTP Response Code -- //
	fmt.Printf("Writing status code ... \n")
	InfoLogger.Printf("Writing status code ... \n")
	w.WriteHeader(response.HttpCode)

	// -- Process Response Body -- //
	fmt.Printf("Writing response body ... \n")
	responseText, err = ioutil.ReadFile(response.Body)

	if err != nil {
		fmt.Printf("Error occured while sending response body .. %v\n", err)
		ErrorLogger.Printf("Error occured while sending response body .. %v\n", err)
	} else {
		w.Write(responseText)
		fmt.Printf("Writing Response Body ... \n%v\n", string(responseText))
		InfoLogger.Printf("Writing Response Body ... \n%v\n", string(responseText))
	}

	fmt.Printf("waiting for next request ...\n")
	InfoLogger.Printf("waiting for next request ...\n")
}

func main() {

	if !readArguments() {
		fmt.Printf("Usage:\n\thttp-mock-server -ip <IPAddress> -port <PORT> -config <config-file-name> [-output <output-file-name>] [-verbose <true|false>]\n")
		os.Exit(-1)
	}
	initializeAndStartListening()

}

func readArguments() bool {

	fmt.Printf("Reading command arguments ...\n")

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

	// Initiate Logger
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	if verboseFlag {
		fmt.Printf("ip\t\t:%v\nport\t\t:%v\nconfigFile\t:%v\nOoutputFile\t:%v\nverboseFlag\t:%v\n", ip, port, configFile, outputFile, verboseFlag)
		InfoLogger.Printf("Command Arguments::\n\tip\t\t:%v\n\tport\t\t:%v\n\tconfigFile\t:%v\n\tOoutputFile\t:%v\n\tverboseFlag\t:%v\n", ip, port, configFile, outputFile, verboseFlag)
	}

	return argsLength > 0
}
