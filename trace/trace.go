package trace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gopaddle/sail/trace/dockerUtils"
	listProcess "gopaddle/sail/trace/listProcess"
	startTrace "gopaddle/sail/trace/startTrace"
	cmd "gopaddle/sail/util/cmd"
	context "gopaddle/sail/util/context"
	json_util "gopaddle/sail/util/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"sort"
	"strconv"
	"time"

	//"strconv"
	"strings"
)

type Details struct {
	Osname string `json:"osname"`
	Osver string `json:"osver"`
	Cmd string `json:"cmd"`
	DirList string `json:"dirlist"`
}

func GetList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	keys := r.URL.Query()
	pid := keys.Get("pid")
	cmd := keys.Get("cmd")
	log.Printf("\n===== Process List =====")
	processes := listProcess.ProcessList()
	if pid == "" && cmd == "" {
		json.NewEncoder(w).Encode(processes)
	} else {
		for _, singleProcess := range processes {
			if pid != "" && singleProcess.Pid == pid {
				log.Printf("Pid: %s", singleProcess.Pid)
				json.NewEncoder(w).Encode(singleProcess)
			} else if cmd != "" && strings.Contains(singleProcess.Cmd, cmd) {
				json.NewEncoder(w).Encode(singleProcess)
			}
		}
	}
}

func StartTracing(w http.ResponseWriter, r *http.Request) {
	os_family, os_name, os_ver := cmd.GetOS()
	log.Printf("Possible Docker Image => %s:%s", os_name, os_ver)

	// os details in context
	Osdetails := startTrace.Osdetails{os_name, os_ver}
	OsMarshal, err := json.Marshal(Osdetails)
	if err != nil {
		log.Println("Osdetails json Marshal error")
	}
	OsJSON := json_util.Parse(OsMarshal)
	context.Instance().SetJSON("os_details", OsJSON)


	if os_family != "NA" {
		/* Install required packages */
		startTrace.CheckRequire(os_name)
	} else {
		log.Fatalf("Unknown os_family")
	}
	keys := r.URL.Query()
	pid := keys.Get("pid")
	if pid == "" {
		log.Printf("Pid: %s does not exist", pid)
	} else {
		/* Get Single Process struct */
		process := listProcess.GetOneProcess(pid)

		/* Save process start command */
		context.Instance().Set("proc_start",process.Cmd)

		kill := fmt.Sprintf("kill %s", process.Pid)
		log.Println(kill)
		cmd.ExecuteAsScript(kill,"Process kill failed")
		log.Printf("\nProcess (PID = %s) success", process.Pid)

		/* Get PUT input as json */
		var trace_input startTrace.TraceInput
		trace_json, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error in Json input startTracing()")
		}
		json.Unmarshal(trace_json, &trace_input)

		/* Get trace time */
		trace_time := trace_input.Time
		log.Println(trace_time)

		/* strace */
		strace := fmt.Sprintf("timeout %ds strace -e trace=file -f -o log/trace.log %s", trace_time, process.Cmd)
		fmt.Println(strace)
		ps := cmd.ExecBg(strace)

		/* Network Tracing */
		new_pid := ps.Process.Pid
		processes := listProcess.ProcessList()
		var pid_list []string
		pid_list = append(pid_list, strconv.Itoa(new_pid))
		for _, singleprocess := range processes {
			if pid == singleprocess.PPid {
				pid_list = append(pid_list, singleprocess.Pid)
			}
		}

		network := startTrace.PortList(trace_time, pid_list)
		network_marshall, err := json.Marshal(network)
		if err != nil{
			log.Println("Json Marshall failed")
		}

		network_json := json_util.Parse(network_marshall)
		context.Instance().SetJSON("network",network_json)

		err = ps.Wait()
		if err != nil {
			log.Println(err)
		}

		time.Sleep(time.Duration(trace_time)*time.Second)

		os_map := context.Instance().GetJSON("os_details")
		os_string := os_map.ToString()
		os_details := startTrace.Osdetails{}
		json.Unmarshal([]byte(os_string), &os_details)

		log.Println("File and Package list making")
		file_list := startTrace.GetDependFiles()
		pkg_list := startTrace.GetDependPackages(os_details.Osname, file_list)

		sort.Strings(file_list)

		/* Packages */
		file, err := os.OpenFile("packages.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		datawriter := bufio.NewWriter(file)

		for _, pkg:= range pkg_list{
			_, _ = datawriter.WriteString(pkg)
		}

		datawriter.Flush()
		file.Close()

		/* Files */
		file, err = os.OpenFile("files.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		datawriter = bufio.NewWriter(file)

		for _, file := range file_list{
			_, _ = datawriter.WriteString(file + "\n")
		}

		datawriter.Flush()
		file.Close()
	}
}

func GetPorts(w http.ResponseWriter, r *http.Request) {
	network_json := context.Instance().GetJSON("network")
	network_string := network_json.ToString()

	var network startTrace.Network
	json.Unmarshal([]byte(network_string), &network)

	json.NewEncoder(w).Encode(network)
}

func GetFilesPkg(w http.ResponseWriter, r *http.Request) {
	// files
	file, err := os.Open("files.log")
	var files []string
	if err != nil {
		log.Println("trace.GetFiles Error : file open failed")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		files = append(files, line)
	}

	// packages
	file, err = os.Open("packages.log")
	var pkg []string
	if err != nil {
		log.Println("trace.GetFiles Error : file open failed")
	}
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		pkg = append(pkg, line)
	}

	// json response
	filepkg := startTrace.FilesPkg{
		Files: files,
		Pkg: pkg,
	}

	json.NewEncoder(w).Encode(filepkg)
}

func NfsMounts(w http.ResponseWriter, r *http.Request) {
	nfs_list := startTrace.GetNfsMounts()
	json.NewEncoder(w).Encode(nfs_list)
}

func GetEnvVariables(w http.ResponseWriter, r *http.Request) {
	env_list := startTrace.GetEnv()
	env_marshall, err := json.Marshal(env_list)
	if err != nil {
		log.Println("Json Marshal failed")
	}

	env_json := json_util.Parse(env_marshall)
	context.Instance().SetJSON("env_list", env_json)

	json.NewEncoder(w).Encode(env_list)
}

func GetShell(w http.ResponseWriter, r *http.Request) {
	shell := startTrace.GetShell()
	json.NewEncoder(w).Encode(shell)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	user := startTrace.GetUser()
	json.NewEncoder(w).Encode(user)
}

func GetStartCmd(w http.ResponseWriter, r *http.Request) {
	start := startTrace.GetStartCmd()
	json.NewEncoder(w).Encode(start)
}

func DockerCreate(w http.ResponseWriter, r *http.Request) {
	var os_details startTrace.Osdetails
	os_det_json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error in Json input startTracing.Osdetails")
	}
	json.Unmarshal(os_det_json, &os_details)
	//
	fmt.Println(os_det_json)
	fmt.Println(os_details)
	//
	os_details = dockerUtils.CheckDockerImage(os_details)

	if (startTrace.Osdetails{}) != os_details {
		log.Println("\ntrace.dockerUtils.Docker:")
		log.Println(os_details)

		dockerUtils.DockerCleanup("dev")
		dockerUtils.DockerCleanup("final")

		dockerUtils.CreateDevImage(os_details)
	} else {
		json.NewEncoder(w).Encode("{\"error\":\"give correct docker image name and version\"}")
	}
}

func DockerCopy(w http.ResponseWriter, r *http.Request) {
	/* Copy User Defined Files */
	var dir_list dockerUtils.DirList
	dir_json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("trace.DockerCopy Error : json read failed")
	}
	json.Unmarshal(dir_json, &dir_list)
	log.Println("List: ",dir_list)

	for _, filename := range dir_list.Dirs {
		dockerUtils.CompressCopy(filename)
	}

	/* Copy all files used by process */
	dockerUtils.CopyProcessFiles()
}

func FinalImageCreate(w http.ResponseWriter, r *http.Request) {
	/* User name */
	user, err := user.Current()
	if err != nil {
		log.Println("trace.FinalImageCreate Error : username retrive error")
	}
	fmt.Println(user)
	/* Working directory */
	//var workdir dockerUtils.Homedir
	//workdir_json, err := ioutil.ReadAll(r.Body)



	/*Osdetails := startTrace.Osdetails{os_name, os_ver}
	OsMarshal, err := json.Marshal(Osdetails)
	if err != nil {
		log.Println("Osdetails json Marshal error")
	}
	OsJSON := json_util.Parse(OsMarshal)
	context.Instance().SetJSON("os_details", OsJSON)*/


	/*var os_details startTrace.Osdetails
	os_det_json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error in Json input startTracing.Osdetails")
	}
	json.Unmarshal(os_det_json, &os_details)*/

	//var image startTrace.Imagename


	//////Imagename := startTrace.Imagename{finalImageName,workdir}
	var imagevar startTrace.Imagename 
	image_json, err := ioutil.ReadAll(r.Body)
	fmt.Println(image_json)
	if err != nil {
		log.Println("trace.FinalImageCreate Error : json read failed")
	}
	//json.Unmarshal(workdir_json, &workdir)
	json.Unmarshal(image_json, &imagevar)
	fmt.Println("Happening")
	fmt.Print(imagevar)
	dockerUtils.FinalImage(user.Username, imagevar.Workdir, imagevar.Finalimagename)
	//dockerUtils.DockerCleanup("dev")
}
