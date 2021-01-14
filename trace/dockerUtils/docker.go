package dockerUtils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gopaddle-io/sail/trace/startTrace"
	"github.com/gopaddle-io/sail/util/cmd"
	"github.com/gopaddle-io/sail/util/context"

	"github.com/sirupsen/logrus"
)

type DirList struct {
	Dirs []string `json:"dirs"`
}

func finished(files []string, line string) bool {
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

func CheckDockerImage(os_put startTrace.Osdetails, slog *logrus.Entry) (startTrace.Osdetails, error) {
	slog.Println("dockerUtils.CheckDockerImage:")
	os_map := context.Instance().GetJSON("os_details")
	os_string := os_map.ToString()
	os_details := startTrace.Osdetails{}
	json.Unmarshal([]byte(os_string), &os_details)

	command := fmt.Sprintf("docker search %s:%s | wc -l", os_details.Osname, os_details.Osver)
	lines_str, err := cmd.ExecuteAsScript(command, "trace.dockerUtils Error: docker search")
	if err != nil {
		return startTrace.Osdetails{}, err
	}
	lines_str = strings.Trim(lines_str, " \n")
	lines, err := strconv.Atoi(lines_str)
	if err != nil {
		slog.Println("trace.dockerutils Error: String conversion failed")
		return startTrace.Osdetails{}, err
	}

	if lines < 2 {
		command := fmt.Sprintf("docker search %s:%s | wc -l", os_put.Osname, os_put.Osver)
		lines_str, _ := cmd.ExecuteAsScript(command, "trace.dockerUtils Error: docker search")
		lines_str = strings.Trim(lines_str, " \n")
		lines, err := strconv.Atoi(lines_str)
		if err != nil {
			slog.Println("trace.dockerutils Error: String conversion failed")
			return startTrace.Osdetails{}, err
		}

		if lines < 2 {
			return startTrace.Osdetails{}, nil
		}
		return os_put, nil
	}
	return os_details, nil
}

func DockerCleanup(container string, slog *logrus.Entry) error {
	//command := fmt.Sprintf("docker stop %s", container)
	//_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker container does not exist")

	command := fmt.Sprintf("docker rm %s", container)
	if _, err := cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker container remove failed"); err != nil {
		return err
	}

	command = fmt.Sprintf("docker rmi %s", container)
	if _, err := cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker image remove failed"); err != nil {
		return err
	}
	return nil
}

func CreateDevImage(os_details startTrace.Osdetails, slog *logrus.Entry, pid string) error {
	home := os.Getenv("HOME")
	file, err := ioutil.ReadFile(home + "/.sail/" + pid + "/packages.log")
	if err != nil {
		slog.Println("Failed on Reading env file : %s ", err.Error())
		return err
	}
	dst := string(file[:])
	p, err := os.Create("./packages.log")
	if err != nil {
		slog.Println("trace.dockerUtils Error: file creation")
		return err
	}
	defer p.Close()

	_, err1 := p.WriteString(dst)
	if err1 != nil {
		slog.Println("trace.dockerUtils Error: file write")
		return err
	}
	dockerfile := fmt.Sprintf(`FROM %s:%s
COPY /packages.log /packages.log
COPY pkg_install.sh /pkg_install.sh

RUN chmod +x /pkg_install.sh && bash /pkg_install.sh

ENV LD_LIBRARY_PATH="/usr/local/lib"

CMD /bin/bash`, os_details.Osname, os_details.Osver)
	f, err := os.Create("./Dockerfile")
	if err != nil {
		slog.Println("trace.dockerUtils Error: file creation")
		return err
	}
	defer f.Close()

	_, err1 = f.WriteString(dockerfile)
	if err1 != nil {
		slog.Println("trace.dockerUtils Error: file write")
		return err
	}

	slog.Println("Container dev build starting...")
	if _, err = cmd.ExecuteAsScript("docker build -t dev .", "trace.dockerUtils Error: image \"dev\" build failed"); err != nil {
		return err
	}
	if _, err = cmd.ExecuteAsScript("docker create -it --name dev dev", "trace.dockerUtils Error: container \"dev\" build failed"); err != nil {
		return err
	}
	if _, err = cmd.ExecuteAsScript("docker start dev", "trace.dockerUtils Error: container \"dev\" start failed"); err != nil {
		// return err
	}
	return nil
}

func CompressCopy(filename string, slog *logrus.Entry) error {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		slog.Print("there is nothing to copy.")
		return err
	}

	if info.IsDir() {
		compress := fmt.Sprintf("tar -czf folder.tar.gz -C %s .", filename)
		if _, err = cmd.ExecuteAsScript(compress, "trace.dockerUtils Error : compression failed"); err != nil {
			// return err
		}
		if _, err = cmd.ExecuteAsScript("docker cp folder.tar.gz dev:folder.tar.gz", "trace.dockerUtils Error : docker copy failed"); err != nil {
			// return err
		}
		zip_remove := fmt.Sprintf("docker exec -i dev bash -c \"mkdir -p %s 2>/dev/null && tar -xzf folder.tar.gz -C %s && rm folder.tar.gz\"", filename, filename)
		if _, err = cmd.ExecuteAsScript(zip_remove, "trace.dockerUtils Error : docker folder.tar.gz copy or extraction failed"); err != nil {
			// return err
		}
	} else {
		copy_file := fmt.Sprintf("docker cp %s dev:%s", filename, filename)
		if _, err = cmd.ExecuteAsScript(copy_file, "trace.dockerUtils Error : docker file copy failed"); err != nil {
			// return err
		}
	}
	return nil
}

func CopyProcessFiles(slog *logrus.Entry) error {
	/* Read Lines */
	file, err := os.Open("files.log")
	var files, files_done []string
	if err != nil {
		slog.Println("trace.dockerUtils Error : file open error")
		return err
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
		line, err = filepath.Abs(line)
		line = strings.TrimRight(line, "/")
		info, err := os.Lstat(line)
		slog.Printf(line)
		if os.IsNotExist(err) {
			slog.Println("trace.dockerUtils Error : file does not exist")
		} else if finished(files_done, line) && !strings.Contains(line, "/lib") {
			if info.IsDir() {
				command := fmt.Sprintf("docker exec dev bash -c \"mkdir -p %s 2>/dev/null\" </dev/null", line)
				if _, err = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker mkdir failed"); err != nil {
					return err
				}
			} else if info.Mode()&os.ModeSymlink != 0 {
				original, err := filepath.EvalSymlinks(line)
				if err != nil {
					log.Println("trace.dockerUtils Error : symlink does not exist")
					return err
				}
				slog.Println("Symlink:")
				slog.Println("Original: ", original)
				dir := filepath.Dir(original)
				command := fmt.Sprintf("docker exec dev bash -c \"mkdir -p %s\" </dev/null", dir)
				slog.Println(command)
				if _, err = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker mkdir failed"); err != nil {
					return err
				}

				command = fmt.Sprintf("docker cp %s dev:%s", original, original)
				slog.Println(command)
				if _, err = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker cp failed"); err != nil {
					return err
				}

				// docker exec rm
				command = fmt.Sprintf("docker exec dev bash -c \"rm -rf %s\" </dev/null", line)
				slog.Println(command)
				output, err := cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker rm failed")
				if err != nil {
					return err
				}

				// docker exec ln
				command = fmt.Sprintf("docker exec dev bash -c \"ln -s %s %s\" </dev/null", original, line)
				slog.Println(command)
				output, err = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker symlink failed")
				if err != nil {
					return err
				}
				slog.Println(output)

				files_done = append(files_done, line)
				files_done = append(files_done, original)
			} else {
				command := fmt.Sprintf("docker cp %s dev:%s", line, line)
				if _, err = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker cp failed"); err != nil {
					return err
				}
				files_done = append(files_done, line)
			}
		}
	}

	// ldconfig to fix libc issue
	command := "docker exec dev bash -c \"ldconfig\" </dev/null"
	slog.Println(command)
	if _, err = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker ldconfig failed"); err != nil {
		return err
	}
	return nil
}

func env_profile(slog *logrus.Entry) error {
	env_json := context.Instance().GetJSON("env_list")
	env_string := env_json.ToString()

	var env_list startTrace.EnvList
	json.Unmarshal([]byte(env_string), &env_list)
	slog.Println(env_list)
	file, err := os.OpenFile("env_list.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Println("trace.dockerUtils Error : file open error")
		return err
	}

	datawriter := bufio.NewWriter(file)
	for _, env := range env_list.Env {
		_, _ = datawriter.WriteString("export " + env.Name + "=" + env.Value + "\n")
	}

	datawriter.Flush()
	file.Close()
	return nil
}

func FinalImage(user string, workdir string, imagename string, slog *logrus.Entry) error {
	if _, err := cmd.ExecuteAsScript("docker stop dev", "trace.dockerUtils Error : docker dev stop failed"); err != nil {
		return nil
	}
	if _, err := cmd.ExecuteAsScript("docker commit dev dev", "trace.dockerUtils Error : docker dev commit failed"); err != nil {
		return nil
	}

	proc_start := context.Instance().Get("proc_start")
	if len(workdir) > 1 {
		workdir = "WORKDIR " + workdir
	}

	// Write environment variables to file
	if err := env_profile(slog); err != nil {
		return err
	}
	dockerfile := fmt.Sprintf(`FROM dev:latest
	
	USER %s

	%s

	COPY env_list.log /home/%s/.profile

	CMD %s && /bin/bash`, user, workdir, user, proc_start)

	f, err := os.Create("./Dockerfile")
	if err != nil {
		slog.Println("trace.dockerUtils Error: file creation")
		return err
	}
	defer f.Close()

	_, err1 := f.WriteString(dockerfile)
	if err1 != nil {
		slog.Println("trace.dockerUtils Error: file write")
		return err
	}

	if _, err = cmd.ExecuteAsScript("docker build -t "+imagename+" .", "trace.dockerUtils Error : docker final build failed"); err != nil {
		// return err
	}
	if _, err = cmd.ExecuteAsScript("docker create -it --name "+imagename+" "+imagename, "trace.dockerUtils Error : docker final create failed"); err != nil {
		// return err
	}
	if _, err = cmd.ExecuteAsScript("docker start "+imagename, "trace.dockerUtils Error : docker final start failed"); err != nil {
		// return err
	}
	if _, err = cmd.ExecuteAsScript("docker commit "+imagename+" "+imagename, "trace.dockerUtils Error : docker final commit failed"); err != nil {
		// return err
	}
	return nil
}
