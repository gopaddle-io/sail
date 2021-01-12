package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"gopaddle/sail/trace/dockerUtils"
	listProcess "gopaddle/sail/trace/listProcess"
	startTrace "gopaddle/sail/trace/startTrace"
	cmd "gopaddle/sail/util/cmd"
	context "gopaddle/sail/util/context"
	json_util "gopaddle/sail/util/json"
	log "gopaddle/sail/util/log"
	"os"
	"os/user"
	"sort"
	"strconv"
	"time"
)

func StartTracing_noreq(pid string, trace_time int, requestID string) (string, error) {
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	os_family, os_name, os_ver, err := cmd.GetOS()
	if err != nil {
		return "", err
	}
	sCxt.Log.Printf("Possible Docker Image : %s:%s", os_name, os_ver)

	// os details in context
	Osdetails := startTrace.Osdetails{os_name, os_ver, ""}
	OsMarshal, err := json.Marshal(Osdetails)
	if err != nil {
		sCxt.Log.Println("Osdetails json Marshal error :", err.Error())
		return "", err
	}
	OsJSON := json_util.Parse(OsMarshal)
	context.Instance().SetJSON("os_details", OsJSON)

	if os_family != "NA" {
		/* Install required packages */
		err := startTrace.CheckRequire(os_name, sCxt.Log)
		if err != nil {
			return "", err
		}
	} else {
		sCxt.Log.Println("Unknown os_family")
		return "", errors.New("Unknown os_family")
	}
	if pid == "" {
		sCxt.Log.Printf("Pid: %s does not exist", pid)
		return "", errors.New("Pid: " + pid + " does not exist")
	} else {
		/* Get Single Process struct */
		process, e := listProcess.GetOneProcess(pid, sCxt.Log)
		if e != nil {
			return "", e
		}
		if process.Pid == "" {
			sCxt.Log.Printf("Pid: %s does not exist", pid)
			return "", errors.New("Pid: " + pid + " does not exist")
		}

		if _, pidDirerr := cmd.ExecuteAsScript("cd ~/.sail && mkdir "+pid, "pid directory creation failed"); pidDirerr != nil {
			return "", pidDirerr
		}

		if e := startTrace.ENVList(pid, sCxt.Log); e != nil {
			return "", e
		}

		network, err := startTrace.PortList(trace_time, pid, sCxt.Log)
		if err != nil {
			return "", err
		}
		network_marshall, err := json.Marshal(network)
		if err != nil {
			sCxt.Log.Println("Json Marshall failed:", err)
			return "", err
		}

		network_json := json_util.Parse(network_marshall)
		context.Instance().SetJSON("network", network_json)

		/* Save process start command */
		context.Instance().Set("proc_start", process.Cmd)
		cwdCmd := fmt.Sprintf("readlink -e /proc/%s/cwd", process.Pid)

		pcwd, err := cmd.ExecuteAsScript(cwdCmd, "Getting Process current working directory failed")
		if err != nil {
			return "", err
		}

		kill := fmt.Sprintf("kill %s", process.Pid)
		sCxt.Log.Println(kill)
		if _, err := cmd.ExecuteAsScript(kill, "process kill failed"); err != nil {
			return "", err
		}
		sCxt.Log.Printf("\nProcess (PID = %s) success", process.Pid)

		/* strace */
		strace := fmt.Sprintf("cd %s && timeout %ds strace -e trace=file -f -o ~/.sail/%s/trace.log %s", pcwd, trace_time, process.Pid, process.Cmd)
		sCxt.Log.Println(strace)
		ps, err := cmd.ExecBg(strace)
		if err != nil {
			return "", err
		}

		/* Network Tracing */
		new_pid := ps.Process.Pid

		processes, err := listProcess.ProcessList(sCxt.Log)
		if err != nil {
			return "", err
		}
		var pid_list []string
		pid_list = append(pid_list, strconv.Itoa(new_pid))
		for _, singleprocess := range processes {
			if pid == singleprocess.PPid {
				pid_list = append(pid_list, singleprocess.Pid)
			}
		}

		err = ps.Wait()
		if err != nil {
			log.Println(err)
		}

		time.Sleep(time.Duration(trace_time) * time.Second)

		os_map := context.Instance().GetJSON("os_details")
		os_string := os_map.ToString()
		os_details := startTrace.Osdetails{}
		json.Unmarshal([]byte(os_string), &os_details)

		sCxt.Log.Println("File and Package list making")
		file_list := startTrace.GetDependFiles(pid, sCxt.Log)
		pkg_list := startTrace.GetDependPackages(os_details.Osname, file_list, sCxt.Log)

		sort.Strings(file_list)

		/* Packages */
		file, err := os.OpenFile("packages.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			sCxt.Log.Printf("failed creating file: %s", err)
			return "", err
		}

		datawriter := bufio.NewWriter(file)

		for _, pkg := range pkg_list {
			_, _ = datawriter.WriteString(pkg)
		}

		datawriter.Flush()
		file.Close()

		/* Files */
		file, err = os.OpenFile("files.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			sCxt.Log.Printf("failed creating file: %s", err)
			return "", err
		}

		datawriter = bufio.NewWriter(file)

		for _, file := range file_list {
			_, _ = datawriter.WriteString(file + "\n")
		}

		datawriter.Flush()
		file.Close()
	}
	return "Strace successfully completed", nil
}

func DockerCreate_noreq(osname string, osver string, imagename string, requestID string) (string, error) {
	var os_details startTrace.Osdetails
	os_details.Osname = osname
	os_details.Osver = osver
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	var err error
	os_details, err = dockerUtils.CheckDockerImage(os_details, sCxt.Log)
	if err != nil {
		return "", err
	}

	if (startTrace.Osdetails{}) != os_details {
		sCxt.Log.Println("\ntrace.dockerUtils.Docker:")
		sCxt.Log.Println(os_details)

		if err := dockerUtils.DockerCleanup("dev", sCxt.Log); err != nil {
			return "", err
		}
		if err := dockerUtils.DockerCleanup(imagename, sCxt.Log); err != nil {
			return "", err
		}

		if err := dockerUtils.CreateDevImage(os_details, sCxt.Log); err != nil {
			return "", err
		}
	} else {
		sCxt.Log.Println("Please give correct OS name and version")
		return "", errors.New("Please give correct OS name and version")
	}
	return "Docker container Build completed successfully", nil
}

func DockerCopy_noreq(dirs []string, requestID string) (string, error) {
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	/* Copy User Defined Files */
	var dir_list dockerUtils.DirList
	dir_list.Dirs = dirs
	sCxt.Log.Println(dirs)
	sCxt.Log.Println("List: ", dir_list)
	sCxt.Log.Println(dir_list)
	sCxt.Log.Println("docker compress initiated")
	for _, filename := range dir_list.Dirs {
		if err := dockerUtils.CompressCopy(filename, sCxt.Log); err != nil {
			return "", err
		}
	}

	/* Copy all files used by process */
	if err := dockerUtils.CopyProcessFiles(sCxt.Log); err != nil {
		return "", err
	}
	return "Docker copy files completed successfully", nil
}

func FinalImageCreate_noreq(workdir string, imagename string, requestID string) (string, error) {
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	/* User name */
	user, err := user.Current()
	if err != nil {
		sCxt.Log.Println("trace.FinalImageCreate Error : username retrive error")
		return "", err
	}
	fmt.Println(user)
	var imagevar startTrace.Imagename
	imagevar.Workdir = workdir
	imagevar.Finalimagename = imagename
	sCxt.Log.Print(user.Username, imagevar.Finalimagename, imagevar.Workdir)
	if err := dockerUtils.FinalImage(user.Username, imagevar.Workdir, imagevar.Finalimagename, sCxt.Log); err != nil {
		return "", err
	}
	return "final docker image copied successfully", nil
}
