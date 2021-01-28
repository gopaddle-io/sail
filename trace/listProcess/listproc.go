package listProcess

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	cmd "github.com/gopaddle-io/sail/util/cmd"

	"github.com/sirupsen/logrus"
)

type Process struct {
	Pid   string `json:"pid"`
	Cmd   string `json:"cmd"`
	PPid  string `json:"ppid"`
	Uid   string `json:"uid"`
	Gid   string `json:"gid"`
	Etime string `json:"time"`
}

func ListLog(vbmode bool) error {
	var args = []string{"-aeo", "uid,gid,pid,ppid,etimes,command"}
	if pidDirerr := cmd.ExecuteAsCommand("cd ~/ && mkdir "+".sail", "sail directory creation failed", vbmode); pidDirerr != nil {
		// return "", pidDirerr
	}
	if pidDirerr := cmd.ExecuteAsCommand("cd ~/.sail && mkdir "+"log", "log directory creation failed", vbmode); pidDirerr != nil {
		// return pidDirerr
	}
	home := os.Getenv("HOME")
	if err := cmd.ExecuteWithOut("ps", vbmode, args, home+"/.sail/log/process_list.log"); err != nil {
		return err
	}
	return nil
}

func ProcessList(slog *logrus.Entry, vbmode bool) ([]Process, error) {
	var processes = []Process{}

	// Execute ps -eo
	if e := ListLog(vbmode); e != nil {
		return processes, e
	}

	user_current, err := user.Current()
	if err != nil {
		if vbmode {
			slog.Println("trace.startTrace Error : GetUser error %s", err.Error())
		}
		return processes, err
	}
	home := os.Getenv("HOME")
	path := fmt.Sprintf("%s/.sail/log/process_list.log", home)
	file, err := os.Open(path)
	if err != nil {
		if vbmode {
			slog.Println("~/.sail/process_list.go file error: %s", err.Error())
		}
		return processes, err
	} else {
		if vbmode {
			slog.Println(" Log file opened ")
		}
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var ps_line = scanner.Text()
		var ps_line_slice = strings.Fields(ps_line)

		if user_current.Uid == ps_line_slice[0] {

			newProcess := Process{
				Pid:   ps_line_slice[2],
				Cmd:   strings.Join(ps_line_slice[5:], " "),
				PPid:  ps_line_slice[3],
				Uid:   ps_line_slice[0],
				Gid:   ps_line_slice[1],
				Etime: ps_line_slice[4],
			}
			processes = append(processes, newProcess)
		}
	}
	// for _, singleProcess := range processes {
	// 	if singleProcess.Pid == "869" {
	// 		fmt.Println("singleProcess.Pid ===>>>", singleProcess.Pid)
	// 	}
	// }
	file.Close()
	// processes = append(processes[:0], processes[1:]...)
	return processes, nil
}

func GetOneProcess(pid string, slog *logrus.Entry, vbmode bool) (Process, error) {
	var process Process
	processes, err := ProcessList(slog, vbmode)
	if err != nil {
		return process, err
	}
	for _, singleProcess := range processes {
		if pid == singleProcess.Pid {
			// str, _ := json.Marshal(singleProcess)
			return singleProcess, nil
		}
	}

	return Process{}, nil
}
