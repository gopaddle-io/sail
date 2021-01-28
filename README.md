[![StackShare](http://img.shields.io/badge/tech-stack-0690fa.svg?style=flat)](https://stackshare.io/gopaddleio/gopaddle)

# sail
Migrate Linux Processes to Docker Containers

# Prerequisite
```sh
sudo groupadd docker
sudo usermod -aG docker <current-user-name>
```
Restart is required after adding user to docker group
## 1. Process List API
```
URL : /api/{accountID}/v1/listProcesses?cmd=nvim
Method : GET
Response :
{
    "pid": "93167",
    "cmd": "nvim ALACRITTY_LOG=/tmp/Alacritty-93120.log COLORTERM=truecolor DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus DESKTOP_SESSION=i3-with-shmlog DISPLAY=:0 GDMSESSION=i3-with-shmlog GDM_LANG=en_US.UTF-8 GTK_MODULES=canberra-gtk-module HOME=/home/kryptiksage I3SOCK=/run/user/1000/i3/ipc-socket.1000 LANG=en_US.UTF-8 LOGNAME=kryptiksage MAIL=/var/spool/mail/kryptiksage MOTD_SHOWN=pam PATH=/home/kryptiksage/.cargo/bin/:/home/kryptiksage/.scripts/:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/bin/site_perl:/usr/bin/vendor_perl:/usr/bin/core_perl PWD=/home/kryptiksage SHELL=/bin/zsh SHLVL=1 TERM=alacritty USER=kryptiksage USERNAME=kryptiksage WINDOWID=60817410 WINDOWPATH=2 XAUTHORITY=/run/user/1000/gdm/Xauthority XDG_DATA_DIRS=/home/kryptiksage/.local/share/flatpak/exports/share:/var/lib/flatpak/exports/share:/usr/local/share:/usr/share XDG_RUNTIME_DIR=/run/user/1000 XDG_SEAT=seat0 XDG_SESSION_CLASS=user XDG_SESSION_DESKTOP=i3-with-shmlog XDG_SESSION_ID=3 XDG_SESSION_TYPE=x11 XDG_VTNR=2 OLDPWD=/home/kryptiksage TERMINAL=alacritty P9K_TTY=old HISTSIZE=10000 SAVEHIST=10000 HISTFILE=/home/kryptiksage/.zsh_history P9K_SSH=0 _=/usr/bin/nvim",
    "ppid": "93125",
    "uid": "1000",
    "gid": "1000",
    "time": "5"
}
```
## 2. Trace API
```
URL : /api/{accountID}/v1/startTracing?pid=93167
Method : PUT
Request :
{
   "time": 2
}
```
## 3. Get files and packages
```
URL : /api/{accountID}/v1/getfilepkg?pid=93167
Method : GET
```
## 4. Get ports
```
URL : /api/{accountID}/v1/getports?pid=93167
Method : GET
```
## 5. Get NFS mounts
```
URL : /api/{accountID}/v1/getNfsMounts?pid=93167
Method : GET
```
## 6. Get Environment variables
```
URL : /api/{accountID}/v1/getenv?pid=93167
Method : GET
```
## 7. Get Default shell
```
URL : /api/{accountID}/v1/getshell?pid=93167
Method : GET
```
## 8. Get UID and GID
```
URL : /api/{accountID}/v1/getuser?pid=93167
Method : GET
```
## 9. Get Start Command
```
URL : /api/{accountID}/v1/getstartcmd?pid=93167
Method : GET
```
## 10. Docker Create dev container
```
URL : /api/{accountID}/v1/dockercreate?pid=93167
Method : PUT
Request(Optional) :
{
   "osname": "archlinux",
   "osver": "latest"
}
```
## 11. Docker Copy Files to dev container
```
URL : /api/{accountID}/v1/dockercopy?pid=93167
Method : PUT
Request :
{
    "dirs": ["testdir", "testfile.txt"]
}
```
## 12. Final docker image creation
```
URL : /api/{accountID}/v1/finalimage?pid=93167
Method : PUT
Request :
{
    "home": "home_directory"
}
```

# Command Line Utility

```
command line utility for sail
```
## To List All Process on Current User
```

$ ./sail list --all process

 1124  1113 /opt/google/chrome/chrome --type=zygote --no-zygote-sandbox
 1125  1113 /opt/google/chrome/chrome --type=zygote
 1126  1124 /opt/google/chrome/nacl_helper
 1127  1125 /opt/google/chrome/nacl_helper
 1130  1125 /opt/google/chrome/chrome --type=zygote
 ```
 
## To Containerize the Process:

## Dockerize Help:
```
$ ./sail dockerize -h


Migrate a running linux process in to a Docker Image. 


sail dockerize --pid <process id> [--time <time in seconds>] [--imageName <docker image name>]

    -p, --pid           pid of the process to trace.
    -t, --time          Time in seconds to trace the process to build its docker profile. Defaults to 2 seconds.
    -i, --imageName     Name of the final docker image. Defaults to 'final'.
    -v, --verbose       Run with Verbose Mode
    -d, --directories   Directories to be copied(seperated by comma)
    
```

## To Strace your process and create a docker image and run as container:

```

$ ./sail dockerize -p 14141 -i nodechecker -t 20 -d /home/bluemeric/codebase/src/gopaddle/nodechecker 


start tracing...
tracing completed
Docker creating...
imageName:  nodechecker
Docker creation completed
Docker file copying ...
Docker file copying completed
Copying fmt file of trace to container...
&{1001 1001 bluemeric  /home/bluemeric}
nodechecker created
```

