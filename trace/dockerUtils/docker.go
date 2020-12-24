package dockerUtils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gopaddle/sail/trace/startTrace"
	"gopaddle/sail/util/cmd"
	"gopaddle/sail/util/context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type DirList struct {
	Dirs []string `json:"dirs"`
}

func finished(files []string, line string) bool{
	for _, file := range files {
		if file == line {
			return false
		}
	}
	return true
}

type Homedir struct {
	Home string `json:"home"`
}

func CheckDockerImage(os_put startTrace.Osdetails) startTrace.Osdetails{
	log.Println("dockerUtils.CheckDockerImage:")
	os_map := context.Instance().GetJSON("os_details")
	os_string :=  os_map.ToString()
	os_details := startTrace.Osdetails{}
	json.Unmarshal([]byte(os_string), &os_details)

	command := fmt.Sprintf("docker search %s:%s | wc -l", os_details.Osname, os_details.Osver)
	lines_str := cmd.ExecuteAsScript(command, "trace.dockerUtils Error: docker search")
	lines_str = strings.Trim(lines_str, " \n")
	lines, err := strconv.Atoi(lines_str)
	if err != nil {
		log.Println("trace.dockerutils Error: String conversion failed")
	}

	if lines < 2 {
		command := fmt.Sprintf("docker search %s:%s | wc -l", os_put.Osname, os_put.Osver)
		lines_str := cmd.ExecuteAsScript(command, "trace.dockerUtils Error: docker search")
		lines_str = strings.Trim(lines_str, " \n")
		lines, err := strconv.Atoi(lines_str)
		if err != nil {
			log.Println("trace.dockerutils Error: String conversion failed")
		}

		if lines < 2 {
			return startTrace.Osdetails{}
		}
		return os_put
	}
	return os_details
}

func DockerCleanup(container string) {
	//command := fmt.Sprintf("docker stop %s", container)
	//_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker container does not exist")

	command := fmt.Sprintf("docker rm %s", container)
	_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker container remove failed")

	command = fmt.Sprintf("docker rmi %s", container)
	_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker image remove failed")
}

func CreateDevImage(os_details startTrace.Osdetails) {
	dockerfile := fmt.Sprintf(`FROM %s:%s
COPY packages.log /packages.log
COPY pkg_install.sh /pkg_install.sh

RUN chmod +x /pkg_install.sh && bash /pkg_install.sh

ENV LD_LIBRARY_PATH="/usr/local/lib"

CMD /bin/bash`, os_details.Osname, os_details.Osver)
	f, err := os.Create("./Dockerfile")
	if err != nil {
		log.Println("trace.dockerUtils Error: file creation")
	}
	defer f.Close()

	_, err1 := f.WriteString(dockerfile)
	if err1 != nil {
		log.Println("trace.dockerUtils Error: file write")
	}

	log.Println("Container dev build starting...")
	_ = cmd.ExecuteAsScript("docker build -t dev .", "trace.dockerUtils Error: image \"dev\" build failed")
	_ = cmd.ExecuteAsScript("docker create -it --name dev dev", "trace.dockerUtils Error: container \"dev\" build failed")
	_ = cmd.ExecuteAsScript("docker start dev", "trace.dockerUtils Error: container \"dev\" start failed")
}

func CompressCopy(filename string) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Fatal("File does not exist.")
	}

	if info.IsDir() {
		compress := fmt.Sprintf("tar -czf folder.tar.gz -C %s .", filename)
		_ = cmd.ExecuteAsScript(compress, "trace.dockerUtils Error : compression failed")
		_ = cmd.ExecuteAsScript("docker cp folder.tar.gz dev:folder.tar.gz", "trace.dockerUtils Error : docker copy failed")
		zip_remove := fmt.Sprintf("docker exec -i dev bash -c \"mkdir -p %s 2>/dev/null && tar -xzf folder.tar.gz -C %s && rm folder.tar.gz\"", filename, filename)
		_ = cmd.ExecuteAsScript(zip_remove, "trace.dockerUtils Error : docker folder.tar.gz copy or extraction failed")
	} else {
		copy_file := fmt.Sprintf("docker cp %s dev:%s", filename, filename)
		_ = cmd.ExecuteAsScript(copy_file,"trace.dockerUtils Error : docker file copy failed")
	}
}

func CopyProcessFiles() {
	/* Read Lines */
	file, err := os.Open("files.log")
	var files, files_done []string
	if err != nil {
		log.Println("trace.dockerUtils Error : file open error")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		files = append(files, line)
	}

	sort.Strings(files)

	/* Copy files to dev container */
	for _, line := range files {
		line,err = filepath.Abs(line)
		line = strings.TrimRight(line, "/")
		info, err := os.Lstat(line)
		log.Printf(line)
		if os.IsNotExist(err) {
			log.Println("trace.dockerUtils Error : file does not exist")
		} else if finished(files_done, line) && !strings.Contains(line, "/lib") {
			if info.IsDir() {
				command := fmt.Sprintf("docker exec dev bash -c \"mkdir -p %s 2>/dev/null\" </dev/null", line)
				_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker mkdir failed")
			} else if info.Mode() & os.ModeSymlink != 0 {
				original, err := filepath.EvalSymlinks(line)
				if err != nil {
					log.Println("trace.dockerUtils Error : symlink does not exist")
				}
				log.Println("Symlink:")
				log.Println("Original: ",original)
				dir := filepath.Dir(original)
				command := fmt.Sprintf("docker exec dev bash -c \"mkdir -p %s\" </dev/null", dir)
				log.Println(command)
				_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker mkdir failed")

				command = fmt.Sprintf("docker cp %s dev:%s", original, original)
				log.Println(command)
				_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker cp failed")

				// docker exec rm
				command = fmt.Sprintf("docker exec dev bash -c \"rm -rf %s\" </dev/null",line)
				log.Println(command)
				output := cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker rm failed")

				// docker exec ln
				command = fmt.Sprintf("docker exec dev bash -c \"ln -s %s %s\" </dev/null", original, line)
				log.Println(command)
				output = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker symlink failed")
				log.Println(output)

				files_done = append(files_done, line)
				files_done = append(files_done, original)
			} else {
				command := fmt.Sprintf("docker cp %s dev:%s", line, line)
				_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker cp failed")
				files_done = append(files_done, line)
			}
		}
	}

	// ldconfig to fix libc issue
	command := "docker exec dev bash -c \"ldconfig\" </dev/null"
	fmt.Println(command)
	_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker ldconfig failed")
}

func env_profile() {
	env_json := context.Instance().GetJSON("env_list")
	env_string := env_json.ToString()

	var env_list startTrace.EnvList
	json.Unmarshal([]byte(env_string), &env_list)
	log.Println(env_list)
	file, err := os.OpenFile("env_list.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("trace.dockerUtils Error : file open error")
	}

	datawriter := bufio.NewWriter(file)
	for _, env := range env_list.Env {
		_, _ = datawriter.WriteString("export "+env.Name+"="+env.Value+"\n")
	}

	datawriter.Flush()
	file.Close()
}

func FinalImage(user string, workdir string, imagename string) {
	_ = cmd.ExecuteAsScript("docker stop dev", "trace.dockerUtils Error : docker dev stop failed")
	_ = cmd.ExecuteAsScript("docker commit dev dev", "trace.dockerUtils Error : docker dev commit failed")

	proc_start := context.Instance().Get("proc_start")
	if len(workdir) > 1 {
		workdir = "WORKDIR " + workdir
	}

	// Write environment variables to file
	env_profile()
	dockerfile := fmt.Sprintf(`FROM dev:latest
	
	USER %s

	%s

	COPY env_list.log /home/%s/.profile

	CMD %s && /bin/bash`, user, workdir, user, proc_start)

	f, err := os.Create("./Dockerfile")
	if err != nil {
		log.Println("trace.dockerUtils Error: file creation")
	}
	defer f.Close()

	_, err1 := f.WriteString(dockerfile)
	if err1 != nil {
		log.Println("trace.dockerUtils Error: file write")
	}


	_ = cmd.ExecuteAsScript("docker build -t " + imagename + " ." , "trace.dockerUtils Error : docker final build failed")
	_ = cmd.ExecuteAsScript("docker create -it --name " + imagename + " " + imagename, "trace.dockerUtils Error : docker final create failed")
	//fmt.Println("create")
	_ = cmd.ExecuteAsScript("docker start " + imagename, "trace.dockerUtils Error : docker final start failed")
	//fmt.Println("Started didnt fail")
	_ = cmd.ExecuteAsScript("docker commit "+imagename + " " + imagename, "trace.dockerUtils Error : docker final commit failed")
	//fmt.Println("AGAIN")
}
