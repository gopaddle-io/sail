package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"sort"
	"strconv"
	"time"

	"github.com/gopaddle-io/sail/trace/dockerUtils"
	listProcess "github.com/gopaddle-io/sail/trace/listProcess"
	startTrace "github.com/gopaddle-io/sail/trace/startTrace"
	cmd "github.com/gopaddle-io/sail/util/cmd"
	context "github.com/gopaddle-io/sail/util/context"
	json_util "github.com/gopaddle-io/sail/util/json"
	clog "github.com/gopaddle-io/sail/util/log"
	log "github.com/gopaddle-io/sail/util/log"
)

func GetList_noreq(requestID string) ([]listProcess.Process, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	processes, err := listProcess.ProcessList(sCxt.Log, true)
	if err != nil {
		return processes, err
	}
	return processes, nil
}
func GetPorts_noreq(requestID string) (startTrace.Network, error) {
	network_json := context.Instance().GetJSON("network")
	network_string := network_json.ToString()

	var network startTrace.Network
	if e := json.Unmarshal([]byte(network_string), &network); e != nil {
		return network, e
	}
	return network, nil
}
func GetFilesPkg_noreq(pid, requestID string) (startTrace.FilesPkg, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	file, err := os.Open("~/.sail/" + pid + "/files.log")
	var files []string
	if err != nil {
		sCxt.Log.Println("module:sail", "requestID:"+requestID)
		return startTrace.FilesPkg{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		files = append(files, line)
	}

	// packages
	home := os.Getenv("HOME")
	file, err = os.Open(home + "/.sail/" + pid + "/packages.log")
	var pkg []string
	if err != nil {
		sCxt.Log.Println("module:sail", "requestID:"+requestID)
		return startTrace.FilesPkg{}, err
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
		Pkg:   pkg,
	}
	return filepkg, nil
}

func NfsMounts_noreq(requestID string) ([]startTrace.Nfs, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	nfs_list, err := startTrace.GetNfsMounts(sCxt.Log)
	if err != nil {
		return nfs_list, err
	}
	return nfs_list, nil
}
func GetEnvVariables_noreq(pid, requestID string) (startTrace.EnvList, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	env_list, err := startTrace.GetEnv(pid, sCxt.Log)
	if err != nil {
		return env_list, err
	}
	return env_list, nil
}
func GetShell_noreq() startTrace.Shell {
	return startTrace.GetShell()
}
func GetUser_noreq() startTrace.User {
	return startTrace.GetUser()
}
func GetStartCmd_noreq() startTrace.Start {
	return startTrace.GetStartCmd()
}
func StartTracing_noreq(pid string, trace_time int, requestID string, vbmode bool) (string, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	os_family, os_name, os_ver, err := cmd.GetOS(vbmode)
	if err != nil {
		return "", err
	}
	if vbmode {
		sCxt.Log.Printf("Possible Docker Image : %s:%s", os_name, os_ver)
	}

	// os details in context
	Osdetails := startTrace.Osdetails{os_name, os_ver, ""}
	OsMarshal, err := json.Marshal(Osdetails)
	if err != nil {
		if vbmode {
			sCxt.Log.Println("Osdetails json Marshal error :", err.Error())
		}
		return "", err
	}
	OsJSON := json_util.Parse(OsMarshal)
	context.Instance().SetJSON("os_details", OsJSON)

	if os_family != "NA" {
		/* Install required packages */
		err := startTrace.CheckRequire(os_name, sCxt.Log, vbmode)
		if err != nil {
			return "", err
		}
	} else {
		if vbmode {
			sCxt.Log.Println("Unknown os_family")
		}
		return "", errors.New("Unknown os_family")
	}
	if pid == "" {
		if vbmode {
			sCxt.Log.Printf("Pid: %s does not exist", pid)
		}
		return "", errors.New("Pid: " + pid + " does not exist")
	} else {
		/* Get Single Process struct */
		process, e := listProcess.GetOneProcess(pid, sCxt.Log, vbmode)
		if e != nil {
			return "", e
		}

		if process.Pid == "" {
			if vbmode {
				sCxt.Log.Printf("Pid: %s does not exist", pid)
			}
			return "", errors.New("Pid: " + pid + " does not exist")
		}
		os.Setenv(pid+"-uid", process.Uid)
		os.Setenv(pid+"-gid", process.Gid)

		if pidDirerr := cmd.ExecuteAsCommand("cd ~/ && mkdir "+".sail", "sail directory creation failed", vbmode); pidDirerr != nil {
			// return "", pidDirerr
		}
		if pidDirerr := cmd.ExecuteAsCommand("cd ~/.sail && mkdir "+pid, "pid directory creation failed", vbmode); pidDirerr != nil {
			// return "", pidDirerr
		}

		if e := startTrace.ENVList(pid, sCxt.Log, vbmode); e != nil {
			return "", e
		}

		network, err := startTrace.PortList(trace_time, pid, sCxt.Log, vbmode)
		if err != nil {
			return "", err
		}
		network_marshall, err := json.Marshal(network)
		if err != nil {
			if vbmode {
				sCxt.Log.Println("Json Marshall failed:", err)
			}
			return "", err
		}

		network_json := json_util.Parse(network_marshall)
		context.Instance().SetJSON("network", network_json)

		/* Save process start command */
		context.Instance().Set("proc_start", process.Cmd)
		cwdCmd := fmt.Sprintf("readlink -e /proc/%s/cwd", process.Pid)

		pcwd, err := cmd.ExecuteAsScript(cwdCmd, "Getting Process current working directory failed", vbmode)
		if err != nil {
			return "", err
		}
		os.Setenv(pid+"-pcwd", pcwd)
		kill := fmt.Sprintf("kill %s", process.Pid)
		if vbmode {
			sCxt.Log.Println(kill)
		}
		if err := cmd.ExecuteAsCommand(kill, "process kill failed", vbmode); err != nil {
			return "", err
		}
		if vbmode {
			sCxt.Log.Printf("\nProcess (PID = %s) success", process.Pid)
		}

		/* strace */
		strace := fmt.Sprintf("timeout %ds strace -e trace=file -f -o ~/.sail/%s/trace.log %s", trace_time, process.Pid, process.Cmd)
		strace = "cd " + pcwd + strace
		if vbmode {
			sCxt.Log.Println(strace)
		}
		ps, err := cmd.ExecBg(strace, vbmode)
		if err != nil {
			return "", err
		}

		/* Network Tracing */
		new_pid := ps.Process.Pid

		processes, err := listProcess.ProcessList(sCxt.Log, vbmode)
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

		ps.Wait()

		time.Sleep(time.Duration(trace_time) * time.Second)

		os_map := context.Instance().GetJSON("os_details")
		os_string := os_map.ToString()
		os_details := startTrace.Osdetails{}
		json.Unmarshal([]byte(os_string), &os_details)
		if vbmode {
			sCxt.Log.Println("File and Package list making")
		}
		file_list, err := startTrace.GetDependFiles(pid, sCxt.Log)
		if err != nil {
			return "", err
		}
		pkg_list := startTrace.GetDependPackages(os_details.Osname, file_list, sCxt.Log, vbmode)

		sort.Strings(file_list)

		/* Packages */
		home := os.Getenv("HOME")
		file, err := os.OpenFile(home+"/.sail/"+pid+"/packages.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if vbmode {
				sCxt.Log.Printf("failed creating file: %s", err)
			}
			return "", err
		}

		datawriter := bufio.NewWriter(file)

		for _, pkg := range pkg_list {
			_, _ = datawriter.WriteString(pkg)
		}

		datawriter.Flush()
		file.Close()

		/* Files */

		file, err = os.OpenFile(home+"/.sail/"+pid+"/files.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if vbmode {
				sCxt.Log.Printf("failed creating file: %s", err)
			}
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

func DockerCreate_noreq(osname string, osver string, imagename string, requestID string, pid string, vbmode bool) (string, error) {
	var os_details startTrace.Osdetails
	os_details.Osname = osname
	os_details.Osver = osver
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	var err error
	os_details, err = dockerUtils.CheckDockerImage(os_details, sCxt.Log, vbmode)
	if err != nil {
		return "", err
	}

	if (startTrace.Osdetails{}) != os_details {
		if vbmode {
			sCxt.Log.Println("\ntrace.dockerUtils.Docker:")
			sCxt.Log.Println(os_details)
		}

		if err := dockerUtils.DockerCleanup("dev", sCxt.Log, vbmode); err != nil {
			// return "", err
		}
		if err := dockerUtils.DockerCleanup(imagename, sCxt.Log, vbmode); err != nil {
			// return "", err
		}

		if err := dockerUtils.CreateDevImage(os_details, sCxt.Log, pid, vbmode); err != nil {
			return "", err
		}
	} else {
		if vbmode {
			sCxt.Log.Println("Please give correct OS name and version")
		}
		return "", errors.New("Please give correct OS name and version")
	}
	return "Docker container Build completed successfully", nil
}

func DockerCopy_noreq(dirs []string, pid, requestID string, vbmode bool) (string, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	/* Copy User Defined Files */
	var dir_list dockerUtils.DirList
	dir_list.Dirs = dirs
	if vbmode {
		sCxt.Log.Println(dirs)
		sCxt.Log.Println("List: ", dir_list)
		sCxt.Log.Println(dir_list)
		sCxt.Log.Println("docker compress initiated")
	}
	for _, filename := range dir_list.Dirs {
		if err := dockerUtils.CompressCopy(filename, sCxt.Log, vbmode); err != nil {
			return "", err
		}
	}

	/* Copy all files used by process */
	if err := dockerUtils.CopyProcessFiles(sCxt.Log, pid, vbmode); err != nil {
		return "", err
	}
	return "Docker copy files completed successfully", nil
}

func FinalImageCreate_noreq(workdir string, imagename string, pid, requestID string, vbmode bool) (string, error) {
	clog.Init()
	log := log.Log("module:sail", "requestID:"+requestID)
	sCxt := NewSailContext(log, requestID)
	/* User name */
	user, err := user.Current()
	if err != nil {
		if vbmode {
			sCxt.Log.Println("trace.FinalImageCreate Error : username retrive error")
		}
		return "", err
	}
	fmt.Println(user)
	var imagevar startTrace.Imagename
	imagevar.Workdir = workdir
	imagevar.Finalimagename = imagename
	if vbmode {
		sCxt.Log.Print(user.Username, imagevar.Finalimagename, imagevar.Workdir)
	}
	if err := dockerUtils.FinalImage(user.Username, imagevar.Workdir, pid, imagevar.Finalimagename, sCxt.Log, vbmode); err != nil {
		return "", err
	}
	return "final docker image copied successfully", nil
}
