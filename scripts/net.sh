#!/bin/bash

#setting options to filter desired protocol
opt="-unp"
prot="udp"

#while loop to get ss stdout filtered based on prot
export x=1
{
while [ $x -lt 3 ]
do
#ingrsip=ingress_ip egrsip=egress_ip
#count=counter ingrsport_ingress_port egrsport=egress_port
	export ingrsip=()
	export ingrsport=()
	export egrsip=()
	export egrsport=()
	export count=0

#filtering ss command with PID entered as stdin
#replacing all occurances of * with % as * is a special character
	ss $opt | grep "pid=$1" | tr -s ' ' | sed  's/\*/%/g' |
	{
#parsing through all stdout to get required ingress and egress info
		while IFS= read -r line
		do
           	     echo $line | cut -d ' ' -f 4 > i
               	     check=$(cat i)
#check contains ingrsip info 
#checking if ingress ip = * as format of ss stdout is different in that case
			if [ "$check" = "%" ]
			then
				ingrsip[$count]=$(echo "*")
				ingrsport[$count]=$(echo $line | cut -d ' ' -f 5)
				egrsip[$count]=$(echo "*")
				egrsport[$count]=$(echo $line | cut -d ' ' -f 7)

#checking if first charecter of ip name is / as the ss stdout format is different then
        	        elif [ ${check:0:1} = "/" ]
        	        then
				ingrsip[$count]=$(echo $line | cut -d ' ' -f 4)
        	                ingrsport[$count]=$(echo $line | cut -d ' ' -f 5)
				egrsip[$count]=$(echo $line | cut -d ' ' -f 6)
				egrsport[$count]=$(echo $line | cut -d ' ' -f 7)
			else
				ingrsip[$count]=$(cut -d ':' -f 1 i)
				ingrsport[$count]=$(cut -d ':' -f 2 i)
				echo $line | cut -d ' ' -f 5 > o
				egrsip[$count]=$(cut -d ':' -f 1 o)
				egrsport[$count]=$(cut -d ':' -f 2 o)
			fi
			count=$((count+1))
		done

#printing in .json format and storing in file
		ingrscount=0
		egrscount=0
		count=$((count-1))
		if [ $count != 0 ]
		then
#logic to check for start of ingress block for printing in .json format
			if [ $x = 1 ]
			then
			echo -e "{\n\tingress:[" > netingrs.json
			fi

			while [ $ingrscount -le $count ]
			do
				echo -e "\t\t{\n\t\t\tip:\"${ingrsip[$ingrscount]}\",\n\t\t\tport:\"${ingrsport[$ingrscount]}\",\n\t\t\tprotocol:\"$prot\"" >> netingrs.json
#logic to check for end of inner ingressinfo block
				if [ $ingrscount = $count ]
				then
					echo -e "\t\t}" >> netingrs.json
				else
					echo -e "\t\t}," >> netingrs.json
				fi
				ingrscount=$((ingrscount+1))
			done
#logic to check for start of egress block (i.e end of ingress block)
			if [ $x = 1 ]
			then
			echo -e "\t],\n\tegress:[" > netegrs.json
			fi

			while [ $egrscount -le $count ]
			do
				echo -e "\t\t{\n\t\t\tip:\"${egrsip[$egrscount]}\",\n\t\t\tport:\"${egrsport[$egrscount]}\",\n\t\t\tprotocol:\"$prot\"" >> netegrs.json
#logic to check for end of inner egressinfo block
				if [ $egrscount = $count ]
				then
					echo -e "\t\t}" >> netegrs.json
				else
					echo -e "\t\t}," >> netegrs.json
				fi

				egrscount=$((egrscount+1))
			done
			if [ $x = 2 ]
			then
			echo -e "\t]\n}" >> netegrs.json
			fi
		elif [ $x = 1 ]
		then
			echo -e "{\n\tingress:[" > netingrs.json
			echo -e "\t],\n\tegress:[" > netegrs.json
		else
			echo -e "\t],\n}" >> netegrs.json
		fi
	}
#chainging filter conditions for ss command from type UDP to TCP
prot="tcp"
opt="-ntp"
x=$((x+1))
done
}

#merging ingress file and egress file in order
cat netegrs.json >> netingrs.json
cat netingrs.json > netinfo.json


