package trace

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	misc "github.com/gopaddle-io/sail/misc"
	"github.com/gopaddle-io/sail/trace/dockerUtils"
	listProcess "github.com/gopaddle-io/sail/trace/listProcess"
	startTrace "github.com/gopaddle-io/sail/trace/startTrace"
	util "github.com/gopaddle-io/sail/util"
	context "github.com/gopaddle-io/sail/util/context"
	json_util "github.com/gopaddle-io/sail/util/json"
	log "github.com/gopaddle-io/sail/util/log"

	"strings"

	"github.com/gorilla/mux"
)

type Details struct {
	Osname  string `json:"osname"`
	Osver   string `json:"osver"`
	Cmd     string `json:"cmd"`
	DirList string `json:"dirlist"`
}

func GetList(rw http.ResponseWriter, req *http.Request) {
	// w.Header().Set("Content-Type", "application/json")
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	slog := log.Log(accID, "module:sail", "requestID:"+requestID)
	slog.Infoln("Requested to Get List")
	keys := req.URL.Query()
	pid := keys.Get("pid")
	cmd := keys.Get("cmd")
	processes, err := listProcess.ProcessList(slog)
	if err != nil {
		response := misc.Response{Code: 500, Response: misc.BuildHTTPErrorJSON(err.Error(), requestID)}
		rw.WriteHeader(response.Code)
		rw.Write([]byte(response.Response))
	}
	if pid == "" && cmd == "" && err != nil {
		json.NewEncoder(rw).Encode(processes)
	}
	for _, singleProcess := range processes {
		if pid != "" && singleProcess.Pid == pid {
			log.Printf("Pid: %s", singleProcess.Pid)
			json.NewEncoder(rw).Encode(singleProcess)
		} else if cmd != "" && strings.Contains(singleProcess.Cmd, cmd) {
			json.NewEncoder(rw).Encode(singleProcess)
		}
	}

}

func StartTracing(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to start a trace")
	keys := req.URL.Query()
	pid := keys.Get("pid")
	var trace_input startTrace.TraceInput
	trace_json, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error in Json input startTracing()")
	}
	json.Unmarshal(trace_json, &trace_input)
	/* Get trace time */
	trace_time := trace_input.Time
	if trace_time == 0 {
		trace_time = 2
	}
	resp, err := StartTracing_noreq(pid, trace_time, requestID)
	if err != nil {
		response := misc.Response{Code: 500, Response: misc.BuildHTTPErrorJSON(err.Error(), requestID)}
		rw.WriteHeader(response.Code)
		rw.Write([]byte(response.Response))
	} else {
		response := misc.Response{Code: 200, Response: misc.BuildHTTPErrorJSON(resp, requestID)}
		rw.WriteHeader(response.Code)
		rw.Write([]byte(response.Response))
	}
}

func GetPorts(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Get Ports")
	network_json := context.Instance().GetJSON("network")
	network_string := network_json.ToString()

	var network startTrace.Network
	json.Unmarshal([]byte(network_string), &network)

	json.NewEncoder(rw).Encode(network)
}

func GetFilesPkg(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to get file packages")
	// files
	keys := req.URL.Query()
	pid := keys.Get("pid")
	file, err := os.Open("~/.sail/" + pid + "/files.log")
	var files []string
	if err != nil {
		log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("trace.GetFiles Error : file open failed")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		files = append(files, line)
	}

	// packages
	file, err = os.Open("~/.sail/" + pid + "/packages.log")
	var pkg []string
	if err != nil {
		log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("trace.GetFiles Error : file open failed")
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

	json.NewEncoder(rw).Encode(filepkg)
}

func NfsMounts(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	slog := log.Log(accID, "module:sail", "requestID:"+requestID)
	slog.Infoln("Requested to get NFS Mounts")
	nfs_list, _ := startTrace.GetNfsMounts(slog)
	json.NewEncoder(rw).Encode(nfs_list)
}

func GetEnvVariables(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	slog := log.Log(accID, "module:sail", "requestID:"+requestID)
	slog.Infoln("Requested to Get ENV variables")
	keys := req.URL.Query()
	pid := keys.Get("pid")
	env_list, _ := startTrace.GetEnv(pid, slog)
	env_marshall, err := json.Marshal(env_list)
	if err != nil {
		log.Println("Json Marshal failed")
	}

	env_json := json_util.Parse(env_marshall)
	context.Instance().SetJSON("env_list", env_json)

	json.NewEncoder(rw).Encode(env_list)
}

func GetShell(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Get Shell")
	shell := startTrace.GetShell()
	json.NewEncoder(rw).Encode(shell)
}

func GetUser(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Get User")
	user := startTrace.GetUser()
	json.NewEncoder(rw).Encode(user)
}

func GetStartCmd(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Get start command")
	start := startTrace.GetStartCmd()
	json.NewEncoder(rw).Encode(start)
}

func DockerCreate(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Docker")
	var os_details startTrace.Osdetails
	os_det_json, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Error in Json input startTracing.Osdetails")
	}
	json.Unmarshal(os_det_json, &os_details)
	DockerCreate_noreq(os_details.Osname, os_details.Osver, os_details.Imagename, requestID)

}

func DockerCopy(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Docker Copy")
	/* Copy User Defined Files */
	var dir_list dockerUtils.DirList
	dir_json, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("trace.DockerCopy Error : json read failed")
	}
	json.Unmarshal(dir_json, &dir_list)
	log.Println("List: ", dir_list)

	DockerCopy_noreq(dir_list.Dirs, requestID)
}

func FinalImageCreate(rw http.ResponseWriter, req *http.Request) {
	requestID := util.NewRequestID()
	defer func() {
		if r := recover(); r != nil {
			e := misc.PanicHandler(r, requestID)
			rw.WriteHeader(e.Code)
			rw.Write([]byte(e.Response))
		}
	}()
	vars := mux.Vars(req)
	accID := vars["accountID"]
	log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("Requested to Get ENV variables")
	/* User name */
	// user, err := user.Current()
	// if err != nil {
	// 	log.Println("trace.FinalImageCreate Error : username retrive error")
	// }
	var imagevar startTrace.Imagename
	image_json, err := ioutil.ReadAll(req.Body)
	log.Println(image_json)
	if err != nil {
		log.Log(accID, "module:sail", "requestID:"+requestID).Infoln("trace.FinalImageCreate Error : json read failed")
	}
	json.Unmarshal(image_json, &imagevar)
	FinalImageCreate_noreq(imagevar.Workdir, imagevar.Finalimagename, requestID)
}
