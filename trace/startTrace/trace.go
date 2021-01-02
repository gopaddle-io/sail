package startTrace

import (
	"bufio"
	"fmt"
	"gopaddle/sail/util/cmd"
	"gopaddle/sail/util/context"
	"log"
	"os"
	"os/user"
	"strings"
)

type Osdetails struct {
	Osname string `json:"osname"`
	Osver string `json:"osver"`
}

type TraceInput struct {
	Time int `json:"time"`
}

type FilesPkg struct {
	Files []string `json:"files"`
	Pkg []string `json:"pkg"`
}

type Network struct {
	Net []Ports `json:"net"`
}

type Ports struct {
	Local Port `json:"local"`
	Peer Port `json:"peer"`
}

type Port struct {
	IP string `json:"ip"`
	Port string `json:"port"`
}

type Nfs struct {
	ServerIP string `json:"serverTime"`
	ClientDir string `json:"clientDir"`
	NfsVer string `json:"nfsVer"`
	Param string `json:"param"`
}

type EnvList struct {
	Env []Env `json:"env"`
}

type Env struct {
	Name string `json:"name"`
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
	Workdir string `json:"workdir"`
}

func CheckRequire(os_name string) {
	switch os_name {
	case "ubuntu" :
		_ = cmd.ExecuteAsScript("dpkg -l strace &>/dev/null || sudo apt install strace", "strace could not be installed")
	case "archlinux" :
		_ = cmd.ExecuteAsScript("pacman -Q strace &>/dev/null || sudo pacman -s strace", "strace could not be installed")
	case "centoslinux" :
		_ = cmd.ExecuteAsScript("rpm -q strace &>/dev/null || sudo yum install strace", "strace could not be installed")
	default:
		log.Fatalf("Unknown OS: %s", os_name)
	}
}

func PortList(delay int, pid_list []string) Network{
	var connect []Ports
	for _, singlepid := range pid_list {
		fmt.Println(singlepid)
		command := fmt.Sprintf("ss -ntp | grep %s > ports.log", singlepid)

		fmt.Println(command)

		file, err := os.Open("ports.log")
		if err != nil {
			log.Println("ports.log open error")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan(){
			line := strings.Split(scanner.Text(), " ")
			log.Println(line)

			local := strings.Split(line[3], ":")
			peer := strings.Split(line[5], ":")

			local_port := Port{
				IP: local[0],
				Port: local[1],
			}
			peer_port := Port{
				IP: peer[0],
				Port: peer[1],
			}

			network := Ports{
				Local: local_port,
				Peer: peer_port,
			}
			connect = append(connect, network)
		}
	}

	networks := Network{
		Net: connect,
	}

	return networks
}

/* Edit trace.log and get file list */

func GetDependFiles() []string {
	var trace_files []string
	file, err := os.Open("./log/trace.log")
	if err != nil {
		log.Printf("trace/startTrace Error: File Open error")
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
	return trace_files
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
func GetDependPackages(os_name string, trace_files []string) []string {
	var pkg_list []string
	log.Println("Os Name: ", os_name)
	for file := range trace_files {
		var pkg_cmd string
		switch os_name {
		case "ubuntu" :
			pkg_cmd = fmt.Sprintf("dpkg -S %s 2>/dev/null | sed 's/[:].*$//g'", trace_files[file])
		case "archlinux" :
			pkg_cmd = fmt.Sprintf("pacman -Qo %s 2>/dev/null | awk '{print $5\"=\"$6}'", trace_files[file])
		case "centoslinux" :
			pkg_cmd = fmt.Sprintf("rpm -qf %s 2>/dev/null", trace_files[file])
		default:
			log.Fatalf("startTrace.GetDependPackages: Unknown OS: %s", os_name)
		}
		pkg_tmp := strings.Split(cmd.ExecuteAsScript(pkg_cmd, "startTrace.GetDependPackages Error"), " ")
		for pkg := range pkg_tmp {
			if !searchSlice(pkg_list, pkg_tmp[pkg]) {
				pkg_list = append(pkg_list, pkg_tmp[pkg])
			}
		}
	}
	return pkg_list
}

func GetNfsMounts() []Nfs{
	file, err := os.Open("/proc/mounts")
	var nfs_list []Nfs
	if err != nil {
		log.Fatalf("trace.NfsMounts Error /proc/mounts error")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
		if fields[2] == "nfs4" || fields[2] == "nfs3" {
			nfs := Nfs {
				ServerIP: fields[0],
				ClientDir: fields[1],
				NfsVer: fields[2],
				Param: fields[3],
			}
			nfs_list = append(nfs_list, nfs)
		}
	}
	return nfs_list
}

func GetEnv() EnvList{
	var env_list []Env
	for _, element := range os.Environ() {
		variable := strings.Split(element, "=")
		env := Env{
			Name: variable[0],
			Value: variable[1],
		}
		env_list = append(env_list, env)
	}

	return EnvList{ Env: env_list }
}

func GetShell() Shell{
	shell := Shell{
		Shell: os.Getenv("SHELL"),
	}

	return shell
}


func GetUser() User{
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

func GetStartCmd() Start{
	cmd := Start{
		Cmd: context.Instance().Get("proc_start"),
	}

	return cmd
}
