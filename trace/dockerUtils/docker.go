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

func CheckDockerImage(os_put startTrace.Osdetails, slog *logrus.Entry, vbmode bool) (startTrace.Osdetails, error) {
	if vbmode {
		slog.Println("dockerUtils.CheckDockerImage:")
	}
	os_map := context.Instance().GetJSON("os_details")
	os_string := os_map.ToString()
	os_details := startTrace.Osdetails{}
	json.Unmarshal([]byte(os_string), &os_details)

	command := fmt.Sprintf("docker search %s:%s | wc -l", os_details.Osname, os_details.Osver)
	lines_str, err := cmd.ExecuteAsScript(command, "trace.dockerUtils Error: docker search", vbmode)
	if err != nil {
		return startTrace.Osdetails{}, err
	}
	lines_str = strings.Trim(lines_str, " \n")
	lines, err := strconv.Atoi(lines_str)
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerutils Error: String conversion failed")
		}
		return startTrace.Osdetails{}, err
	}
	if lines < 2 {
		command := fmt.Sprintf("docker search %s:%s | wc -l", os_put.Osname, os_put.Osver)
		lines_str, _ := cmd.ExecuteAsScript(command, "trace.dockerUtils Error: docker search", vbmode)
		lines_str = strings.Trim(lines_str, " \n")
		lines, err := strconv.Atoi(lines_str)
		if err != nil {
			if vbmode {
				slog.Println("trace.dockerutils Error: String conversion failed")
			}
			return startTrace.Osdetails{}, err
		}

		if lines < 2 {
			return startTrace.Osdetails{}, nil
		}
		return os_put, nil
	}
	return os_details, nil
}

func DockerCleanup(container string, slog *logrus.Entry, vbmode bool) error {
	//command := fmt.Sprintf("docker stop %s", container)
	//_ = cmd.ExecuteAsScript(command, "trace.dockerUtils Error : docker container does not exist")

	command := fmt.Sprintf("docker rm -f %s", container)
	if err := cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker container remove failed", vbmode); err != nil {
		return err
	}

	command = fmt.Sprintf("docker rmi -f %s", container)
	if err := cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker image remove failed", vbmode); err != nil {
		return err
	}
	return nil
}

func CreateDevImage(os_details startTrace.Osdetails, slog *logrus.Entry, pid string, vbmode bool) error {
	home := os.Getenv("HOME")
	file, err := ioutil.ReadFile(home + "/.sail/" + pid + "/packages.log")
	if err != nil {
		if vbmode {
			slog.Println("Failed on Reading env file : %s ", err.Error())
		}
		return err
	}
	dst := string(file[:])
	p, err := os.Create("./packages.log")
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file creation")
		}
		return err
	}
	defer p.Close()

	_, err1 := p.WriteString(dst)
	if err1 != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file write")
		}
		return err
	}
	dockerfile := fmt.Sprintf(`FROM %s:%s
COPY /packages.log /packages.log
COPY pkg_install.sh /pkg_install.sh

RUN chmod +x /pkg_install.sh && bash /pkg_install.sh

ENV LD_LIBRARY_PATH="/usr/local/lib"

CMD /bin/bash`, os_details.Osname, os_details.Osver)
	e := os.Remove("./Dockerfile")
	if e != nil {
		fmt.Println("error while removing doecker file :", e)
	}
	f, err := os.Create("./Dockerfile")
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file creation")
		}
		return err
	}
	defer f.Close()

	_, err1 = f.WriteString(dockerfile)
	if err1 != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file write")
		}
		return err
	}
	if vbmode {
		slog.Println("Container dev build starting...")
	}
	if err = cmd.ExecuteAsCommand("docker build -t dev .", "trace.dockerUtils Error: image \"dev\" build failed", vbmode); err != nil {
		return err
	}
	if err = cmd.ExecuteAsCommand("docker create -it --name dev dev", "trace.dockerUtils Error: container \"dev\" build failed", vbmode); err != nil {
		return err
	}
	if err = cmd.ExecuteAsCommand("docker start dev", "trace.dockerUtils Error: container \"dev\" start failed", vbmode); err != nil {
		// return err
	}
	return nil
}

func CompressCopy(filename string, slog *logrus.Entry, vbmode bool) error {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if vbmode {
			slog.Print("there is nothing to copy.")
		}
		return err
	}

	if info.IsDir() {
		compress := fmt.Sprintf("tar -czf folder.tar.gz -C %s .", filename)
		if err = cmd.ExecuteAsCommand(compress, "trace.dockerUtils Error : compression failed", vbmode); err != nil {
			// return err
		}
		if err = cmd.ExecuteAsCommand("docker cp folder.tar.gz dev:folder.tar.gz", "trace.dockerUtils Error : docker copy failed", vbmode); err != nil {
			// return err
		}
		zip_remove := fmt.Sprintf("docker exec -i dev bash -c \"mkdir -p %s 2>/dev/null && tar -xzf folder.tar.gz -C %s && rm folder.tar.gz\"", filename, filename)
		if err = cmd.ExecuteAsCommand(zip_remove, "trace.dockerUtils Error : docker folder.tar.gz copy or extraction failed", vbmode); err != nil {
			// return err
		}
	} else {
		copy_file := fmt.Sprintf("docker cp %s dev:%s", filename, filename)
		if err = cmd.ExecuteAsCommand(copy_file, "trace.dockerUtils Error : docker file copy failed", vbmode); err != nil {
			// return err
		}
	}
	return nil
}

func CopyProcessFiles(slog *logrus.Entry, pid string, vbmode bool) error {
	/* Read Lines */
	home := os.Getenv("HOME")
	file, err := os.Open(home + "/.sail/" + pid + "/files.log")
	var files, files_done []string
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error : file open error")
		}
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
		if vbmode {
			slog.Printf(line)
		}
		if os.IsNotExist(err) {
			if vbmode {
				slog.Println("trace.dockerUtils Error : file does not exist")
			}
		} else if finished(files_done, line) && !strings.Contains(line, "/lib") {
			if info != nil && info.IsDir() {
				command := fmt.Sprintf("docker exec dev bash -c \"mkdir -p %s 2>/dev/null\" </dev/null", line)
				if err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker mkdir failed", vbmode); err != nil {
					return err
				}
			} else if info != nil && info.Mode()&os.ModeSymlink != 0 {
				original, err := filepath.EvalSymlinks(line)
				if err != nil {
					if vbmode {
						log.Println("trace.dockerUtils Error : symlink does not exist")
					}
					return err
				}
				if vbmode {
					slog.Println("Symlink:")
					slog.Println("Original: ", original)
				}
				dir := filepath.Dir(original)
				command := fmt.Sprintf("docker exec dev bash -c \"mkdir -p %s\" </dev/null", dir)
				if vbmode {
					slog.Println(command)
				}
				if err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker mkdir failed", vbmode); err != nil {
					return err
				}

				command = fmt.Sprintf("docker cp %s dev:%s", original, original)
				if vbmode {
					slog.Println(command)
				}
				if err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker cp failed", vbmode); err != nil {
					return err
				}

				// docker exec rm
				command = fmt.Sprintf("docker exec dev bash -c \"rm -rf %s\" </dev/null", line)
				if vbmode {
					slog.Println(command)
				}
				err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker rm failed", vbmode)
				if err != nil {
					return err
				}

				// docker exec ln
				command = fmt.Sprintf("docker exec dev bash -c \"ln -s %s %s\" </dev/null", original, line)
				if vbmode {
					slog.Println(command)
				}
				err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker symlink failed", vbmode)
				if err != nil {
					// return err
				}

				files_done = append(files_done, line)
				files_done = append(files_done, original)
			} else {
				command := fmt.Sprintf("docker cp %s dev:%s", line, line)
				if err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker cp failed", vbmode); err != nil {
					// return err
				}
				files_done = append(files_done, line)
			}
		}
	}

	// ldconfig to fix libc issue
	command := "docker exec dev bash -c \"ldconfig\" </dev/null"
	if vbmode {
		slog.Println(command)
	}
	if err = cmd.ExecuteAsCommand(command, "trace.dockerUtils Error : docker ldconfig failed", vbmode); err != nil {
		return err
	}
	return nil
}

func env_profile(slog *logrus.Entry, pid string, vbmode bool) error {
	env_list, err := startTrace.GetEnv(pid, slog)
	if err != nil {
		return err
	}
	home := os.Getenv("HOME")
	file, err := os.OpenFile(home+"/.sail/"+pid+"/listenv.log", os.O_APPEND|os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error : file open error")
		}
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

func FinalImage(user string, workdir string, pid, imagename string, slog *logrus.Entry, vbmode bool) error {
	if err := cmd.ExecuteAsCommand("docker stop dev", "trace.dockerUtils Error : docker dev stop failed", vbmode); err != nil {
		return nil
	}
	if err := cmd.ExecuteAsCommand("docker commit dev dev", "trace.dockerUtils Error : docker dev commit failed", vbmode); err != nil {
		return nil
	}

	proc_start := context.Instance().Get("proc_start")
	if len(workdir) > 1 {
		workdir = "WORKDIR " + workdir
	}

	// Write environment variables to file
	if err := env_profile(slog, pid, vbmode); err != nil {
		return err
	}
	home := os.Getenv("HOME")
	file, err := ioutil.ReadFile(home + "/.sail/" + pid + "/listenv.log")
	if err != nil {
		if vbmode {
			slog.Println("Failed on Reading env file : %s ", err.Error())
		}
		return err
	}
	dst := string(file[:])
	p, err := os.Create("./listenv.log")
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file creation")
		}
		return err
	}
	defer p.Close()

	_, err1 := p.WriteString(dst)
	if err1 != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file write")
		}
		return err1
	}
	uid := os.Getenv(pid + "-uid")
	// gid := os.Getenv(pid + "-gid")
	pcwd := os.Getenv(pid + "-pcwd")
	pcwd = strings.Trim(pcwd, " \n")
	dockerfile := fmt.Sprintf(`FROM dev:latest
	RUN if ! id -u %s > /dev/null 2>&1; then useradd -ms /bin/bash %s --uid=%s ; fi
	
	USER %s

	WORKDIR %s

	COPY listenv.log /home/%s/.profile

	CMD cd %s && %s && /bin/bash`, user, user, uid, user, home, user, pcwd, proc_start)

	//  CMD sleep 3000 && /bin/bash`, user, user, uid, user, home, user)

	f, err := os.Create("./Dockerfile")
	if err != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file creation")
		}
		return err
	}
	defer f.Close()

	_, err2 := f.WriteString(dockerfile)
	if err2 != nil {
		if vbmode {
			slog.Println("trace.dockerUtils Error: file write")
		}
		return err2
	}

	if err = cmd.ExecuteAsCommand("docker build -t "+imagename+" .", "trace.dockerUtils Error : docker final build failed", vbmode); err != nil {
		// return err
	}
	if err = cmd.ExecuteAsCommand("docker create -it --name "+imagename+" "+imagename, "trace.dockerUtils Error : docker final create failed", vbmode); err != nil {
		// return err
	}
	if err = cmd.ExecuteAsCommand("docker start "+imagename, "trace.dockerUtils Error : docker final start failed", vbmode); err != nil {
		// return err
	}
	if err = cmd.ExecuteAsCommand("docker commit "+imagename+" "+imagename, "trace.dockerUtils Error : docker final commit failed", vbmode); err != nil {
		// return err
	}
	return nil
}
