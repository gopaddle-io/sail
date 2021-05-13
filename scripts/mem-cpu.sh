#!/bin/bash

# meml=mem limit memr=mem request
# cpul=cpu limit cpur=cpu request

#filtering total memory available in systema and storing as memory limit
meml=$(cat /proc/meminfo | grep MemTotal | tr -s ' ' | cut -d ' ' -f 2 | bc)
#filtering current memory useage of process using PID entered and storing as memory request
memr=$(pmap $1 | tail -1 | tr -s ' ' | cut -d ' ' -f 2 | tr -d 'K' | bc)

#logic to convert memory values from K to M ( i.e to .json readable format)
meml="$(( meml / 1000 )).$(( meml % 1000 )) M"
memr="$(( memr / 1000 )).$(( memr % 1000 )) M"

#filtering cpu limit from /proc/cpuinfo and converting core to millicore (m)
cpul=$(( $(cat /proc/cpuinfo | grep cores | cut -d ':' -f 2 | tr -d ' ' |tail -1 | bc) * 1000 ))
#filtering cpu request from ps -o %cpu and converting to millicore (m)
cpur=$(echo "$(ps -p $1 -o %cpu | tail -1 | bc )*$(( cpul / 100 ))" | bc)

#printing in .json format
echo -e "\"resources\":{\n\t\"limits\":{\n\t\t\"memory\":\"$meml\",\n\t\t\"cpu\":\"$cpul m\"\n\t}," > mem.cpu.json
echo -e "\t\"requests\":{\n\t\t\"memory\":\"$memr\",\n\t\t\"cpu\":\"$cpur m\"\n\t}\n}" >> mem.cpu.json
