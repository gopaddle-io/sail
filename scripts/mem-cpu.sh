#!/bin/bash

# meml=mem limit memr=mem request
# cpul=cpu limit cpur=cpu request
meml=$(cat /proc/meminfo | grep MemTotal | tr -s ' ' | cut -d ' ' -f 2 | bc)
memr=$(pmap $1 | tail -1 | tr -s ' ' | cut -d ' ' -f 2 | tr -d 'K' | bc)

#converting K to M
meml="$(( meml / 1000 )).$(( meml % 1000 )) M"
memr="$(( memr / 1000 )).$(( memr % 1000 )) M"

#converting cores to millicores
cpul=$(( $(cat /proc/cpuinfo | grep cores | cut -d ':' -f 2 | tr -d ' ' |tail -1 | bc) * 1000 ))
cpur=$(echo "$(ps -p $1 -o %cpu | tail -1 | bc )*$(( cpul / 100 ))" | bc)

#printing in .json format
echo -e "\"resources\":{\n\t\"limits\":{\n\t\t\"memory\":\"$meml\",\n\t\t\"cpu\":\"$cpul m\"\n\t}," > mem-cpu.json
echo -e "\t\"requests\":{\n\t\t\"memory\":\"$memr\",\n\t\t\"cpu\":\"$cpur m\"\n\t}\n}" >> mem-cpu.json
