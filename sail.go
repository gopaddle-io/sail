package main

import (
	directory "gopaddle/sail/directory"
	trace "gopaddle/sail/trace"
	util "gopaddle/sail/util"
	json "gopaddle/sail/util/json"
	clog "gopaddle/sail/util/log"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
	"github.com/gorilla/mux"
)

const (
	config = "./config/profiles-%s.json"
)

func main() {
	port := ":9000"

	runtime.GOMAXPROCS(runtime.NumCPU())
	args := os.Args[1:]
	var env string
	if len(args) < 1 {
		log.Println("Sever Environment is not given,\n\t $ ./server <profile>")
		os.Exit(0)
	}

	env = args[0]
	log.Println("Selected Environment : ", env)

	//Loading Config data to context
	res := directory.LoadConfig(env)
	if res == false {
		log.Println("Config not loaded")
		os.Exit(0)
	}
	clog.Init()
	//To load configDirectory
	go loadDirectoryConfig(env)
	//db.Instance().Info()

	router := mux.NewRouter()
	router.Headers("Content-Type", "application/json", "X-Requested-With", "XMLHttpRequest")

	/* Server Specific API */
	router.HandleFunc("/api/status", GetStatus).Methods("GET")
	router.HandleFunc("/api/version", GetVersion).Methods("GET")

	/* Process List API */
	router.HandleFunc("/api/{accountID}/v1/listProcesses", trace.GetList).Methods("GET")

	/* Trace API */
	router.HandleFunc("/api/{accountID}/v1/startTracing", trace.StartTracing).Methods("PUT")

	/* File and Package list API */
	router.HandleFunc("/api/{accountID}/v1/getfilepkg", trace.GetFilesPkg).Methods("GET")

	/* Ports */
	router.HandleFunc("/api/{accountID}/v1/getports", trace.GetPorts).Methods("GET")

	/* Environment Variables */
	router.HandleFunc("/api/{accountID}/v1/getenv", trace.GetEnvVariables).Methods("GET")

	/* NFS Mount API */
	router.HandleFunc("/api/{accountID}/v1/getNfsMounts", trace.NfsMounts).Methods("GET")

	/* Default shell API */
	router.HandleFunc("/api/{accountID}/v1/getshell", trace.GetShell).Methods("GET")

	/* UID and GID API*/
	router.HandleFunc("/api/{accountID}/v1/getuser", trace.GetUser).Methods("GET")

	/* Start Command API */
	router.HandleFunc("/api/{accountID}/v1/getstartcmd", trace.GetStartCmd).Methods("GET")

	/* Docker Checks */
	router.HandleFunc("/api/{accountID}/v1/dockercreate", trace.DockerCreate).Methods("PUT")

	/* Docker Copy Files and Folders */
	router.HandleFunc("/api/{accountID}/v1/dockercopy", trace.DockerCopy).Methods("PUT")

	/* Create Final Image */
	router.HandleFunc("/api/{accountID}/v1/finalimage", trace.FinalImageCreate).Methods("PUT")

	log.Printf("Server listening at %s", port)
	e := http.ListenAndServe(port, router)
	if e != nil {
		log.Fatalf("Error While establishing connection: %v", e)
	}
}

// GetStatus to get status of service
func GetStatus(rw http.ResponseWriter, req *http.Request) {
	jobj := json.New()
	jobj.Put("status", "Running")
	rw.WriteHeader(200)
	rw.Write([]byte(jobj.ToString()))
}

// GetVersion to get version of service
func GetVersion(rw http.ResponseWriter, req *http.Request) {
	config, err := util.LoadConfig(config, "")
	jobj := json.New()
	if err != nil {
		jobj.Put("reason", err.Error())
		rw.WriteHeader(404)
		rw.Write([]byte(jobj.ToString()))
		return
	}
	jobj.Put("version", config.GetString("version"))
	rw.WriteHeader(200)
	rw.Write([]byte(jobj.ToString()))
}

//It will periodically load configDirectory
func loadDirectoryConfig(env string) {
	for {
		directory.LoadDirectory(env)
		//Wait for 30 secs between each call
		timer := time.NewTimer(time.Second * 300)
		<-timer.C
	}
}
