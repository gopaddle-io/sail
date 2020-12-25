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
	//"io/ioutil"
	"log"
	//"net/http"
	"os"
	//"os/user"
	"sort"
	"strconv"
	"time"

	//"strconv"
	//"strings"
)

/*type Details struct {
	Osname string `json:"osname"`
	Osver string `json:"osver"`
	Cmd string `json:"cmd"`
	DirList string `json:"dirlist"`
}*/



func StartTracing_noreq(pid string, trace_time int) {
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

func DockerCreate_noreq(osname string, osver string, imagename string) {
	var os_details startTrace.Osdetails
	//os_det_json, err := ioutil.ReadAll(r.Body)
	/*if err != nil {
		log.Printf("Error in Json input startTracing.Osdetails")
	}*/
	//json.Unmarshal(os_det_json, &os_details)
	//
	//fmt.Println(os_det_json)
	os_details.Osname = osname
	os_details.Osver = osver
	fmt.Println(os_details)
	//
	os_details = dockerUtils.CheckDockerImage(os_details)

	if (startTrace.Osdetails{}) != os_details {
		log.Println("\ntrace.dockerUtils.Docker:")
		log.Println(os_details)

		dockerUtils.DockerCleanup("dev")
		dockerUtils.DockerCleanup(imagename)

		dockerUtils.CreateDevImage(os_details)
	} else {
		fmt.Println("Please give correct OS name and version")
	}
}

