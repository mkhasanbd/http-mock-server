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

	fmt.Printf("Reading Configuration from: '%v' ...\n", configFile)
	readConfiguration()

	http.HandleFunc("/", defaultHandler)
	fmt.Println("listing on " + ip + ":" + port + " ...\n")
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
		fmt.Printf("Error occured while sending response header .. : %v\n", err)
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

	// -- Process Response Body -- //

	responseText, err = ioutil.ReadFile(response.Body)

	if err != nil {
		fmt.Printf("Error occured while sending response body .. %v\n", err)
		ErrorLogger.Printf("Error occured while sending response body .. \n%v", err)
	} else {
		w.Write(responseText)
	}

	//------- logging code ------//

	if verboseFlag {

		var buff strings.Builder

		fmt.Println("---- Incoming request trace ---- ")
		fmt.Fprintf(&buff, "\n---- Incoming request trace ----\n")

		fmt.Printf("Request URL:: %v\n", req.RequestURI)
		fmt.Printf("HTTP Method:: %v\n", req.Method)

		fmt.Fprintf(&buff, "Request URL:: %v\n", req.RequestURI)
		fmt.Fprintf(&buff, "HTTP Method:: %v\n", req.Method)

		fmt.Println("\nHeader:: ")
		fmt.Fprintf(&buff, "\nHeader:: \n")

		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Printf("\t%v: %v\n", name, h)
				fmt.Fprintf(&buff, "\t%v: %v\n", name, h)
			}
		}

		req.ParseForm()

		fmt.Println("\nrequest.Form::")
		fmt.Fprintf(&buff, "\nrequest.Form::\n")

		for key, value := range req.Form {
			fmt.Printf("\t%s : %s\n", key, value)
			fmt.Fprintf(&buff, "\t%s : %s\n", key, value)
		}

		fmt.Println("\nrequest.PostForm::")
		for key, value := range req.PostForm {
			fmt.Printf("\t%s : %s\n", key, value)
			fmt.Fprintf(&buff, "\t%s : %s\n", key, value)
		}

		bodyStream, err := ioutil.ReadAll(req.Body)

		if err == nil {
			fmt.Println("\nrequest.Body::")
			fmt.Printf("\n%v\n", string(bodyStream))

			fmt.Fprintf(&buff, "\nrequest.Body::")
			fmt.Fprintf(&buff, "\n%v\n", string(bodyStream))

		} else {
			fmt.Printf("Error occured while sending response body .. %v\n", err)
			ErrorLogger.Printf("Error occured while sending response body .. \n%v", err)
		}
		fmt.Println("---------------End---------------- ")
		fmt.Fprintf(&buff, "---------------End---------------- ")
		InfoLogger.Println(buff.String())
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
		fmt.Printf("Supplied Arguments ... \nip :\t%v\nport :\t%v\nconfigFile :\t%v\nOoutputFile :\t%v\nverboseFlag :\t%v\n", ip, port, configFile, outputFile, verboseFlag)
	}

	return argsLength > 0
}
