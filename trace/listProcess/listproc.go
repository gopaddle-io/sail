package listProcess

import (
	"bufio"
	"log"
	"os"
	"strings"
	cmd "gopaddle/sail/util/cmd"
)

type Process struct {
	Pid string `json:"pid"`
	Cmd string `json:"cmd"`
	PPid string `json:"ppid"`
	Uid string `json:"uid"`
	Gid string `json:"gid"`
	Etime string `json:"time"`
}


func ListLog() {
	var args = []string {"-aeo", "uid,gid,pid,ppid,etimes,command"}
	cmd.ExecuteWithOut("ps", args, "./log/process_list.log")
}

func ProcessList() []Process {
	var processes = []Process{}

	// Execute ps -eo
	ListLog()

	file, err := os.Open("./log/process_list.log")
	if err != nil {
		log.Fatalf("util/tools/process_list.go file error: %s", err)
	} else {
		log.Println("===== Log file opened =====")
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var ps_line = scanner.Text()
		var ps_line_slice = strings.Fields(ps_line)
		newProcess := Process {
			Pid: ps_line_slice[2],
			Cmd: strings.Join(ps_line_slice[5:], " "),
			PPid: ps_line_slice[3],
			Uid: ps_line_slice[0],
			Gid: ps_line_slice[1],
			Etime: ps_line_slice[4],
		}
		processes = append(processes, newProcess)
	}
	file.Close()
	processes = append(processes[:0], processes[1:]...)
	return processes
}

func GetOneProcess(pid string) Process {
	processes := ProcessList()
	for _, singleProcess := range processes {
		if pid == singleProcess.Pid {
			return singleProcess
		}
	}
	return Process{}
}
