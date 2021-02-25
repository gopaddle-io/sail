package startTrace

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/gopaddle-io/sail/util/cmd"
	"github.com/gopaddle-io/sail/util/context"
	"github.com/gopaddle-io/sail/util/log"

	"github.com/sirupsen/logrus"
)

type Osdetails struct {
	Osname    string `json:"osname"`
	Osver     string `json:"osver"`
	Imagename string `json:"Imagename"`
}

type TraceInput struct {
	Time int `json:"time"`
}

type FilesPkg struct {
	Files []string `json:"files"`
	Pkg   []string `json:"pkg"`
}

type Network struct {
	Net []Port `json:"net"`
}

type Ports struct {
	Local Port `json:"local"`
	Peer  Port `json:"peer"`
}

type Port struct {
	// IP   string `json:"ip"`
	Port string `json:"port"`
}

type Nfs struct {
	ServerIP  string `json:"serverTime"`
	ClientDir string `json:"clientDir"`
	NfsVer    string `json:"nfsVer"`
	Param     string `json:"param"`
}

type EnvList struct {
	Env []Env `json:"env"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type User struct {
	UID string `json:"uid"`
	GID string `json:"gid"`
}

type Shell struct {
	Shell string `json:"shell"`
}

type Start struct {
	Cmd string `json:"cmd"`
}

type Imagename struct {
	Finalimagename string `json:"finalimagename"`
	Workdir        string `json:"workdir"`
}

func CheckRequire(os_name string, slog *logrus.Entry, vbmode bool) error {
	// log := log.Log("module:sail", "requestID:"+requestID)
	switch os_name {
	case "ubuntu":
		if err := cmd.ExecuteAsCommand("dpkg -l strace &>/dev/null || sudo apt install strace", "strace could not be installed", vbmode); err != nil {
			return err
		}
	case "archlinux":
		if err := cmd.ExecuteAsCommand("pacman -Q strace &>/dev/null || sudo pacman -s strace", "strace could not be installed", vbmode); err != nil {
			return err
		}
	case "centoslinux":
		if err := cmd.ExecuteAsCommand("rpm -q strace &>/dev/null || sudo yum install strace", "strace could not be installed", vbmode); err != nil {
			return err
		}
	default:
		err := fmt.Sprintf("Unknown OS: %s", os_name)
		return errors.New(err)
	}
	return nil
}
func ENVList(pid string, slog *logrus.Entry, vbmode bool) error {
	catcmd := `cat /proc/` + pid + `/environ | tr '\0' ' ' > ~/.sail/` + pid + `/listenv.log`
	if err := cmd.ExecuteAsCommand(catcmd, "env list failed", vbmode); err != nil {
		return err
	}
	return nil
}

func WritePID(delay int, pid string, vbmode bool) {
	//	fmt.Println("==================ENTERING WRITEPID=================   ", pid)
	command := fmt.Sprintf("ps --forest -g $(ps -o sid= -p %s) |awk '{ if ( NR > 2  ) { print } }' | awk '{print $1}' > ~/.sail/%s/pidList.log", pid, pid)
	if err := cmd.ExecuteAsCommand(command, "ports log list failed", vbmode); err != nil {
		fmt.Printf("trace/writePID Error: File Open error")
	}

}

func PortList(delay int, pid string, slog *logrus.Entry, vbmode bool) (Network, error) {
	WritePID(delay, pid, vbmode)
	networks := Network{}
	home := os.Getenv("HOME")
	pidFile, err := os.Open(home + "/.sail/" + pid + "/pidList.log")
	if err != nil {
		if vbmode {
			slog.Println("ports.log open error", err)
		}
		return networks, err
	}
	defer pidFile.Close()
	deleteCommand := fmt.Sprintf("rm ~/.sail/%s/ports.log;touch ~/.sail/%s/ports.log", pid, pid)
	if err := cmd.ExecuteAsCommand(deleteCommand, "ports log delete failed", vbmode); err != nil {
		// return networks, err
	}
	scanner_pid := bufio.NewScanner(pidFile)
	var processes []string
	//scanner := bufio.NewScanner()
	scanner_pid.Split(bufio.ScanLines)
	for scanner_pid.Scan() {
		line := scanner_pid.Text()
		processes = append(processes, line)
	}
	for _, childPid := range processes {
		command := fmt.Sprintf("ss -tulpn | grep %s | awk '{print $5}'| cut -d ':' -f 2 >> ~/.sail/%s/ports.log", childPid, pid)
		// ss -l -p -n | grep pid=10749 > ~/.sail/10749/ports.log
		// fmt.Println(command)
		if err := cmd.ExecuteAsCommand(command, "ports log list failed", vbmode); err != nil {
			// return networks, err
		}
	}
	// home := os.Getenv("HOME")
	file, err := os.Open(home + "/.sail/" + pid + "/ports.log")
	if err != nil {
		if vbmode {
			slog.Println("ports.log open error", err)
		}
		return networks, err
	}
	defer file.Close()
	var Ports []Port
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		//ports := strings.Split(scanner.Text(), ":")
		port := scanner.Text()
		local_port := Port{
			Port: port,
		}
		Ports = append(Ports, local_port)
	}

	networks.Net = Ports

	return networks, nil
}

/* Edit trace.log and get file list */

func GetDependFiles(pid string, slog *logrus.Entry) ([]string, error) {
	var trace_files []string
	home := os.Getenv("HOME")
	file, err := os.Open(home + "/.sail/" + pid + "/trace.log")
	if err != nil {
		slog.Printf("trace/startTrace Error: File Open error")
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Count(line, "\"") != 0 {
			line := line[strings.Index(line, "\"")+1:]
			line = line[:strings.Index(line, "\"")]
			if !(searchSlice(trace_files, line) || strings.Contains(line, "/docker") || strings.Contains(line, "/proc") || strings.Contains(line, "/dev") || strings.Contains(line, "/sys")) {
				trace_files = append(trace_files, line)
			}
		}
	}
	trace_files = append(trace_files, "/etc/group", "/etc/passwd")
	return trace_files, nil
}

/* Check if substr present in slice */
func searchSlice(slice []string, substr string) bool {
	for str := range slice {
		if slice[str] == substr {
			return true
		}
	}
	return false
}

/* Get package dependencies */
func GetDependPackages(os_name string, trace_files []string, slog *logrus.Entry, vbmode bool) []string {
	var pkg_list []string
	if vbmode {
		log.Println("Os Name: ", os_name)
	}
	for file := range trace_files {
		var pkg_cmd string
		switch os_name {
		case "ubuntu":
			pkg_cmd = fmt.Sprintf("dpkg -S %s 2>/dev/null | sed 's/[:].*$//g'", trace_files[file])
		case "archlinux":
			pkg_cmd = fmt.Sprintf("pacman -Qo %s 2>/dev/null | awk '{print $5\"=\"$6}'", trace_files[file])
		case "centoslinux":
			pkg_cmd = fmt.Sprintf("rpm -qf %s 2>/dev/null", trace_files[file])
		default:
			slog.Printf("startTrace.GetDependPackages: Unknown OS: %s", os_name)

		}
		str, _ := cmd.ExecuteAsScript(pkg_cmd, "startTrace.GetDependPackages Error", vbmode)
		pkg_tmp := strings.Split(str, " ")
		for pkg := range pkg_tmp {
			if !searchSlice(pkg_list, pkg_tmp[pkg]) {
				pkg_list = append(pkg_list, pkg_tmp[pkg])
			}
		}
	}
	return pkg_list
}

func GetNfsMounts(slog *logrus.Entry) ([]Nfs, error) {
	file, err := os.Open("/proc/mounts")
	var nfs_list []Nfs
	if err != nil {
		slog.Printf("trace.NfsMounts Error /proc/mounts error")
		return nfs_list, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
		if fields[2] == "nfs4" || fields[2] == "nfs3" {
			nfs := Nfs{
				ServerIP:  fields[0],
				ClientDir: fields[1],
				NfsVer:    fields[2],
				Param:     fields[3],
			}
			nfs_list = append(nfs_list, nfs)
		}
	}
	return nfs_list, nil
}

func GetEnv(pid string, slog *logrus.Entry) (EnvList, error) {
	envmap := make(map[string]string)
	home := os.Getenv("HOME")
	file, err := ioutil.ReadFile(home + "/.sail/" + pid + "/listenv.log")
	if err != nil {
		// slog.Println("Failed on Reading env file : %s ", err.Error())
		return EnvList{}, err
	}
	dst := string(file[:])
	envs := strings.Split(dst, " ")
	for _, env := range envs {
		envArr := strings.Split(env, "=")
		if len(envArr) == 2 {
			envmap[envArr[0]] = envArr[1]
		}
	}
	var envList EnvList
	for k, v := range envmap {
		var env Env
		env.Name = k
		env.Value = v
		envList.Env = append(envList.Env, env)
	}
	return envList, nil
}

func GetShell() Shell {
	shell := Shell{
		Shell: os.Getenv("SHELL"),
	}

	return shell
}

func GetUser() User {
	user_current, err := user.Current()
	if err != nil {
		log.Println("trace.startTrace Error : GetUser error")
	}

	user := User{
		UID: user_current.Uid,
		GID: user_current.Gid,
	}

	return user
}

func GetStartCmd() Start {
	cmd := Start{
		Cmd: context.Instance().Get("proc_start"),
	}

	return cmd
}
