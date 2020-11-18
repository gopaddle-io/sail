package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/pborman/uuid"
)

/*

Command from the Template Engine

<cmd[n]>.<linux>.<name>.<version>

<linux>=`cmd1 ls
	 cmm2 mkdir ps`

*/

var Scripts = getBaseScripts()

var linuxScript = `cwd=readlink /proc/%s/cwd
cmd=cat -A /proc/%s/cmdline
exe=readlink /proc/%s/exe
net=lsof -P -p %s -n | grep IPv
net.once=lsof -P -i  | grep LISTEN
link=lsof -P -i :%s -sTCP:LISTEN
service=lsof -i -sTCP:LISTEN`

func CheckAgentCompatibility() bool {
	if os_family, _, _ := GetOS(); strings.Contains(strings.ToLower(os_family), "linux") {
		return true
	} else if strings.Contains(strings.ToLower(os_family), "windows") {
		return false
	} else if strings.Contains(strings.ToLower(os_family), "darwin") {
		return false
	}
	return false
}

func StringToFile(script string) string {
	var fname = fmt.Sprintf("/tmp/%s.sh", uuid.New())
	ioutil.WriteFile(fname, []byte(script), 0755)
	return fname
}

func FileToString(fname string) string {
	data, err := ioutil.ReadFile(fname)
	log.Println("Found ", err)
	return string(data)
}

func Execute(cmd string,err_message string, param string) string {
	ps, err := exec.Command(cmd, param).Output()
	if err != nil {
		log.Fatalf("util/cmd Error: %s", err_message)
		return "Error"
	}
	result := string(ps)
	return result
}

func ExecuteWithOut(cmd string, param []string, filename string) {
	ps := exec.Command(cmd, param...)
	file, _ := os.Create(filename)
	ps.Stdout = file

	err := ps.Start()
	if err != nil {
		log.Printf("util/cmd Error: %s", err)
	} else {
		log.Printf("===== %s Executed =====", cmd)
	}
	ps.Wait()
}

func ExecuteAsScript(cmd string, err_message string, param ...string) string {
	var stderr bytes.Buffer
	fname := StringToFile(cmd)
	params := append([]string{fname}, param...)
	ps := exec.Command("bash", params...)
	ps.Stderr = &stderr
	data, err := ps.Output()
	if err != nil {
		log.Printf("util/cmd ExecuteAsScript: %s : %s", err, err_message)
	}
	os.Remove(fname)
	return string(data)
}

func ExcuteAsScriptOut(cmd string, filename string, param ...string) {
	fname := StringToFile(cmd)
	params := append([]string{fname}, param...)
	ps := exec.Command("bash", params...)
	file, _ :=  os.Create(filename)
	ps.Stdout = file

	err := ps.Start()
	if err != nil {
		log.Printf("util/cmd Error")
	} else {
		log.Printf("===== %s Executed =====", cmd)
	}
	ps.Wait()
}

func ExecBg(cmd string, param ...string) *exec.Cmd{
	fname := StringToFile(cmd)
	params := append([]string{fname}, param...)
	ps := exec.Command("bash", params...)
	err := ps.Start()
	if err != nil {
		log.Printf("util/cmd error ExecBg() command failed")
	}
	return ps
}

func RemoveUnprintable(s string) string {
	re := regexp.MustCompile("\\^@")
	return re.ReplaceAllString(s, " ")
}

func GetOS() (string, string, string) {
	var os_family, os_name, os_ver = "NA", "NA", "latest"
	os := runtime.GOOS
	if strings.Contains(strings.ToLower(os), "linux") {
		tmp_name := ExecuteAsScript("grep \"^NAME\" /etc/os-release | cut -d \"=\" -f 2 | tr -d \"\\\" \\n\" | head -n 1", "")
		tmp_ver:= ExecuteAsScript("grep \"^VERSION_ID\" /etc/os-release | cut -d \"=\" -f 2 | tr -d \"\\\" \\n\"", "")
		if tmp_name != "" {
			os_name = strings.ToLower(tmp_name)
		}
		if tmp_ver != "" {
			os_ver = tmp_ver
		}
		os_family = os

	} else if strings.Contains(strings.ToLower(os), "windows") {
		log.Println("Found Window and Un-Supported")
	} else if strings.Contains(strings.ToLower(os), "darwin") {
		log.Println("Found MacOS and Un-Supported")
	}
	return os_family, os_name, os_ver
}

func GetScript(name string) string {
	//Script cant be empty
	return Scripts[name]
}

func GetScriptf(name string, v ...interface{}) string {
	return fmt.Sprintf(Scripts[name], v...)
}

func GetConfigFile() string {
	return "/etc/mobilizer/agent.conf"
}

func getBaseScripts() map[string]string {
	// log.Printf("Loading Base Scripts for Agent [%s.(%s).(%s)]......", os_family, os_name, os_ver)
	var script string
	if os_family, _, _ := GetOS(); strings.Contains(os_family, "linux") || strings.Contains(os_family, "darwin") {
		script = linuxScript
	}
	return getScripts(script)
}

func getScripts(script string) map[string]string {
	cmd_list := strings.Split(script, "\n")
	m := make(map[string]string)
	for i := 0; i < len(cmd_list); i++ {
		if cmd_sep := strings.Split(cmd_list[i], "="); len(cmd_sep) > 1 {
			key := strings.ToLower(cmd_sep[0])
			cmd := cmd_sep[1]
			m[key] = cmd
		} else {
			log.Println("Unknown Script", cmd_list[i])
		}
	}
	return m
}

func WriteFile(file string, data []byte) error {
	if errWrite := ioutil.WriteFile(file, data, 0644); errWrite != nil {
		return errWrite
	} else {
		return nil
	}
}
