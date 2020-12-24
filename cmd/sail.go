package main

import (
	
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"bytes"
	"os/exec"
	"io/ioutil"
)

func main() {

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	dockerizeCommand := flag.NewFlagSet("dockerize", flag.ExitOnError)
	helpCommand := flag.NewFlagSet("help", flag.ExitOnError)
	listTextPtr := listCommand.String("all", "", "What to list (Required), list --all process") 
	listHelpPtr := listCommand.Bool("help",false,"Help regarding commands")
	dockerizeTextPtr := dockerizeCommand.String("pid", "", "pid of process (Required)")
	dockerizeTimePtr := dockerizeCommand.String("time","2","Time for which trace commands runs")
	dockerizeImagePtr := dockerizeCommand.String("imageName","final","Final image name")
	dockerizeHelpPtr := dockerizeCommand.Bool("help",false,"Help regarding commands")

	if len(os.Args) < 2 {
		fmt.Println("Command required.")
		fmt.Println("To check the commands possible, run sail help")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
	case "dockerize":
		dockerizeCommand.Parse(os.Args[2:]) 
	case "help":
		helpCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if listCommand.Parsed() {

		if *listHelpPtr == true {

			fmt.Print("\nArguments\n")
			fmt.Printf("%-12s%s\n","all","consider all processes (syntax : saillist --all process)")
			os.Exit(0)
		}

		if *listTextPtr == "" {
			listCommand.PrintDefaults()
			os.Exit(1)
		}

		if os.Args[3] == "process" {
			out, err := exec.Command("ps", "-eo", "pid,ppid,cmd").Output()
			if err != nil {
				fmt.Printf("%s", err)
			}
			output := string(out[:])
			fmt.Println(output)
		} else {
			fmt.Println("Enter complete command (try : sail list --all process)")
		}		
	}

	if dockerizeCommand.Parsed() {


		if *dockerizeHelpPtr == true {

			fmt.Print("\nArguments\n")
			fmt.Printf("%-12s%s\n","pid","Process pid (required field)")
			fmt.Printf("%-12s%s\n","time","Time for which the process should be traced")
			fmt.Printf("%-12s%s\n","imageName","Final image name the image created should be stored with (Default imageName is : final)")
			os.Exit(0)
		}

		if *dockerizeTextPtr == "" {
			dockerizeCommand.PrintDefaults()
			os.Exit(1)
		}
		
		values := map[string]string{"time":*dockerizeTimePtr}
		values2 := map[string]string{"osname":"ubuntu", "osver":"20.04"}
		values4 := map[string]string{"finalimagename": *dockerizeImagePtr,"home": "/tmp"}
		dir := [2]string{"packages.log", "pkg_install.sh"}
		values3 := map[string][2]string{"dirs": dir}
		jsonStr, _ := json.Marshal(values)
		url := "http://localhost:9000/api/1/v1/startTracing?pid="+ *dockerizeTextPtr
		url1 := "http://localhost:9000/api/1/v1/dockercreate"
		url2 := "http://localhost:9000/api/1/v1/dockercopy"
		url3 := "http://localhost:9000/api/1/v1/finalimage"
		
		client := &http.Client{}
		json1, err := json.Marshal(values2)
		if err != nil {
			panic(err)
		}
		json2, err3 := json.Marshal(values3)
		if err3 != nil {
			panic(err3)
		}
		json3, err4 := json.Marshal(values4)
		if err4 != nil {
			panic(err4)
		}		
		request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		} else {
			defer response.Body.Close()
			_, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Print("Tracing pid = " + *dockerizeTextPtr)
			fmt.Println("   ", response.StatusCode)
			///////////////////////////////////////////////////////////////////////////////////
			client1 := &http.Client{}
			req, err := http.NewRequest("PUT", url1, bytes.NewBuffer(json1))
			if err != nil {
				fmt.Println(err)
			}
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			resp, err := client1.Do(req)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print("Creating container for image = " + *dockerizeImagePtr)
			fmt.Println("  ",resp.StatusCode)
			////////////////////////////////////////////////////////////////////////////////////
			req1, err1 := http.NewRequest(http.MethodPut, url2, bytes.NewBuffer(json2))
			if err1 != nil {
				fmt.Println(err1)
			}
			req1.Header.Set("Content-Type", "application/json; charset=utf-8")
			resp, err2 := client.Do(req)
			if err2 != nil {
				fmt.Println(err2)
			}
			fmt.Print("Copying log file of trace to container...")
			fmt.Println("   ",resp.StatusCode)
			////////////////////////////////////////////////////////////////////////////////////
			req3, err6 := http.NewRequest(http.MethodPut, url3, bytes.NewBuffer(json3))
			if err6 != nil {
				fmt.Println(err6)
			}
			req1.Header.Set("Content-Type", "application/json; charset=utf-8")
			resp, err5 := client.Do(req3)
			if err5 != nil {
				fmt.Println(err5)
			}
			fmt.Print(*dockerizeImagePtr + " created")
			fmt.Println("   ",resp.StatusCode)
			fmt.Println("To check the image, use command : docker image inspect " + *dockerizeImagePtr)
		}
	}

	if helpCommand.Parsed() {
		
		fmt.Println("Enter sail [command] --help for more details on specific commands and their arguments.\nUse --[argument] for additional arguments")
		fmt.Println("\nCommands")
		fmt.Printf("%-12s%s\n","list","list processes")
		fmt.Printf("%-12s%s\n","dockerize","given a process id (pid),it traces the process, creates a container for it and creates the image")
	}
}

