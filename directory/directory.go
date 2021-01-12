package directory

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gopaddle-io/sail/util/cmd"
	"github.com/gopaddle-io/sail/util/context"
	"github.com/gopaddle-io/sail/util/json"
)

const (
	config      = "./config/profiles-%s.json"
	errorConfig = "./config/error_config.json"
	directory   = "./config/service_directory-%s.json"
)

func Endpoint(svc string, suffix string) string {
	return EndpointWithSSL(svc, suffix, false)
}

func EndpointWithSSL(svc string, suffix string, ssl bool) string {
	//Get host and port if not found default it to localhost , 8769
	var jobj = json.New()
	ep := context.Instance().Get(svc)
	jobj = json.ParseString(ep)
	epHost := jobj.GetString("host")
	epPort := jobj.GetString("port")
	sslPrefix := ""
	if ssl {
		sslPrefix = "s"
	}
	return fmt.Sprintf("http%s://%s:%s/%s", sslPrefix, epHost, epPort, suffix)
}

func ErrorFmt(scope string, err_name string, v ...interface{}) string {
	if val := ErrorString(scope, err_name); strings.Contains(val, "%") {
		return fmt.Sprintf(val, v...)
	} else {
		return val
	}
}

func ErrorString(scope string, name string) string {
	errPool := context.Instance().GetJSON("errors")
	jErr := errPool.GetJSON(scope)
	var value string
	if value = jErr.GetString(name); value == "" {
		log.Println("WARN : error is empty ", name)
		return "Unknown Error"
	}
	return value
}

func LoadConfig(env string) bool {
	var profile = json.New()

	//Reading Configuration file
	file, err := ioutil.ReadFile(fmt.Sprintf(config, env))
	if err != nil {
		log.Fatalf("Configuraiton '%s' file not found", fmt.Sprintf(config, env))
		return false
	}

	// Retrieving Configuration
	data := json.Parse(file)
	profile = data.GetJSON("mongodb")
	context.Instance().SetObject("db-endpoint", profile.GetAsStringArray("db-endpoint"))
	context.Instance().Set("db-port", profile.GetString("db-port"))
	context.Instance().Set("db-name", profile.GetString("db-name"))
	context.Instance().Set("user-db", profile.GetString("user-db"))
	context.Instance().Set("db-user", profile.GetString("db-user"))
	context.Instance().Set("db-password", profile.GetString("db-password"))

	log.Println("Environment: ", env)

	osFamily, osName, osVersion, err := cmd.GetOS()
	if err != nil {
		log.Println("Error while loading Env: ", err)
		os.Exit(0)
	}
	context.Instance().Set("osFamily", osFamily)
	context.Instance().Set("osName", osName)
	context.Instance().Set("osVersion", osVersion)

	log.Printf("DB-IP: %s DB-Name: %s\n", context.Instance().GetObject("db-endpoint"), context.Instance().Get("db-name"))
	return LoadDirectory(env)
}

func LoadDirectory(env string) bool {

	var config = []string{errorConfig, directory}
	for _, v := range config {

		// Env based variables
		for _, filename := range []string{directory} {
			if v == filename {
				v = fmt.Sprintf(filename, env)
			}
		}

		//Reading Configuration file
		file, err := ioutil.ReadFile(v)
		if err != nil {
			log.Printf(" %s Configuraiton file not found: %v", v, err)
			return false
		}

		// Retrieving Configuration
		data := json.Parse(file)

		if strings.Contains(v, "error_config") {
			context.Instance().SetJSON("errors", data)
		}

		if strings.Contains(v, "service_directory") {
			keyList := data.GetKeyList()
			for _, key := range keyList {
				sD := data.GetJSON(key)
				context.Instance().Set(key, sD.ToString())
				fmt.Printf("%s:%s\n", key, context.Instance().Get(key))

			}
		}
	}
	return true
}
