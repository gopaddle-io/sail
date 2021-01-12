package listProcess

import (
	"bufio"
	cmd "gopaddle/sail/util/cmd"
	"os"
	"os/user"
	"strings"

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

func ListLog() error {
	var args = []string{"-aeo", "uid,gid,pid,ppid,etimes,command"}
	if err := cmd.ExecuteWithOut("ps", args, "./log/process_list.log"); err != nil {
		return err
	}
	return nil
}

func ProcessList(slog *logrus.Entry) ([]Process, error) {
	var processes = []Process{}

	// Execute ps -eo
	if e := ListLog(); e != nil {
		return processes, e
	}

	user_current, err := user.Current()
	if err != nil {
		slog.Println("trace.startTrace Error : GetUser error %s", err.Error())
		return processes, err
	}

	file, err := os.Open("./log/process_list.log")
	if err != nil {
		slog.Println("util/tools/process_list.go file error: %s", err.Error())
		return processes, err
	} else {
		slog.Println(" Log file opened ")
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
	file.Close()
	processes = append(processes[:0], processes[1:]...)
	return processes, nil
}

func GetOneProcess(pid string, slog *logrus.Entry) (Process, error) {
	var process Process
	processes, err := ProcessList(slog)
	if err != nil {
		return process, err
	}
	for _, singleProcess := range processes {
		if pid == singleProcess.Pid {
			return singleProcess, nil
		}
	}
	return Process{}, nil
}
