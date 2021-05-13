#!/bin/bash

echo "" > vol-acces.json
echo -e "[" > storageClass.json

# minvol:volume_request  maxvol:volume_limit  accesmod:access_mode
# filepath:file_path  counter:count_variable  mntsource:mount_source
minvol=()
maxvol=()
filepath=()
acsmod=()
mntsource=()
counter=0


#greps all nfs mounted dirs or files
mount | grep "nfs" | cut -d ' ' -f 3 |
	{
#parsing through stdout to get necessary info
 		while IFS= read -r line
		do
#filtering findmnt stdout to get access perms
#filtering fuser stdout to get PID of process accesing mounted file
#storing in file along with moun path
			echo "$line,$(findmnt $line | tail -1 | tr -s ' ' | cut -d ' ' -f 4 | cut -d ',' -f 1):$(fuser -m $line)" >> vol-acces.json
		done

#grepping line containing stdin PID (along with access perms and filepath), from the above mention file
	cat vol-acces.json | grep $1 | cut -d ':' -f 1 |
		{
			while IFS= read -r lin
			do
				filepath[$counter]=$(echo $lin | cut -d ',' -f 1)

#filtering df command stdout to get min & max vol and store in array
				minvol[$counter]=$(df $(echo $lin | cut -d ',' -f 1) | tail -1 | tr -s ' ' | cut -d ' ' -f 3)
				maxvol[$counter]=$(df $(echo $lin | cut -d ',' -f 1) | tail -1 | tr -s ' ' | cut -d ' ' -f 4)
#filtering $line to get feild containing access perms and store in array
				acsmod[$counter]=$(echo $lin | cut -d ',' -f 2)

#converting vol from K to M
				minvol[$counter]="$(( minvol[$counter] / 1000 )).$((minvol[$counter] % 1000))M"
				maxvol[$counter]="$(( maxvol[$counter] / 1000 )).$((maxvol[$counter] % 1000))M"

#converting access perms to .json file format
				if [ ${acsmod[$counter]} == "rw" ]
				then
					acsmod[$counter]="ReadWrite"
				elif [ ${acsmod[$counter]} == "ro" ]
				then
					acsmod[$counter]="ReadOnly"
				fi

#filtering findmt stdout to get mount source of mounted filesystem
				mntsource[$counter]=$(findmnt -m $(echo $lin | cut -d ',' -f 1) | tail -1 | tr -s ' ' | cut -d ' ' -f 2 )
#printing nfsServer and sharePath in .json format by filtering $mntsource
				echo  -e "\t{\n\t\t\"nfsServer\":\"$(echo ${mntsource[$counter]} | cut -d ':' -f 1 )\",\n\t\t\"sharePath\":\"$(echo ${mntsource[$counter]} | cut -d ':' -f 2)\"," >> storageClass.json
				echo -e "\t\t\"fileSystem\":\"nfs4\"" >> storageClass.json
#logic to check for end of block while printing in .json format
				if [ $((counter+1)) = $(cat vol-acces.json | grep $1 | wc -l) ]
				then
					echo -e "\t}" >> storageClass.json
				else
					echo -e "\t}," >> storageClass.json
				fi

				counter=$((counter+1))
			done

#converting to .json format
			count=0
			counter=$((counter - 1))
			echo "" > vol-acces.json
			while [ $count -le $counter ]
			do
				echo -e "[\n\t{\n\t\t\"filePath\":\"${filepath[$count]}\"\n\t\t\"accessMode\":\"${acsmod[$count]}\",\n\t\t\"resources\":{\n\t\t\t\"limits\":\"${maxvol[$count]}\",\n\t\t\t\"requests\":\"${minvol[$count]}\"\n\t\t\t},\n\t\t\"volumeMode\":\"FileSystem\"" >> vol-acces.json
#logic to check for end of block while printing in .hson format
				if [ $count -eq $counter ]
				then
					echo -e "\t}" >> vol-acces.json
				else
					echo -e "\t}," >> vol-acces.json
				fi
				count=$((count + 1))
			done
			echo "]" >> vol-acces.json
			echo "]" >> storageClass.json
		}
	}
