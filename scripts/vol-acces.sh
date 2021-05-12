#!/bin/bash

#creates file to store filepath and the pid using those filepaths
echo "" > vol.acces.json

# minvol:volume_request  maxvol:volume_limit  accesmod:access_mode
# filepath:file_path  counter:count_variable
minvol=()
maxvol=()
filepath=()
acsmod=()
counter=0

#greps all nfs mounted dirs or files
mount | grep "nfs" | cut -d ' ' -f 3 |
	{

#stores filepaths along with PIDs and AccessMode of processes opening the files
 		while IFS= read -r line
		do
			echo "$line,$(findmnt $line | tail -1 | tr -s ' ' | cut -d ' ' -f 4 | cut -d ',' -f 1):$(fuser -m $line)" >> vol.acces.json
		done

#grepping filepath of stdin PID, from file
	cat vol.acces.json | grep $1 | cut -d ':' -f 1 |
		{
			while IFS= read -r lin
			do
				filepath[$counter]=$(echo $lin | cut -d ',' -f 1)

#storing vol limit and request in array
				minvol[$counter]=$(df $(echo $lin | cut -d ',' -f 1) | tail -1 | tr -s ' ' | cut -d ' ' -f 3)
				maxvol[$counter]=$(df $(echo $lin | cut -d ',' -f 1) | tail -1 | tr -s ' ' | cut -d ' ' -f 4)

				acsmod[$counter]=$(echo $lin | cut -d ',' -f 2)

#converting vol from K to M
				minvol[$counter]="$(( minvol[$counter] / 1000 )).$((minvol[$counter] % 1000))M"
				maxvol[$counter]="$(( maxvol[$counter] / 1000 )).$((maxvol[$counter] % 1000))M"

#storing AccessMode in array
				if [ ${acsmod[$counter]} == "rw" ]
				then
					acsmod[$counter]="ReadWrite"
				elif [ ${acsmod[$counter]} == "ro" ]
				then
					acsmod[$counter]="ReadOnly"
				fi
				counter=$((counter+1))
			done

#converting to .json format
			count=0
			counter=$((counter - 1))
			echo "" > vol.acces.json
			while [ $count -le $counter ]
			do
				echo -e "[\n\t{\n\t\t\"filepath\":\"${filepath[$count]}\"\n\t\t\"accessMode\":\"${acsmod[$count]}\",\n\t\t\"resources\":{\n\t\t\t\"limits\":\"${maxvol[$count]}\",\n\t\t\t\"requests\":\"${minvol[$count]}\"\n\t\t\t},\n\t\t\"volumeMode\":\"FileSystem\"" >> vol.acces.json
				if [ $count -eq $counter ]
				then
					echo -e "\t}" >> vol.acces.json
				else
					echo -e "\t}," >> vol.acces.json
				fi
				count=$((count + 1))
			done
			echo "]" >> vol.acces.json

#displaying .json file
			echo -e "\n\n\n\n\n"
			cat vol.acces.json
		}
	}
