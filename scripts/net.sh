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

#replacing all occurances of * with % as * is a special character
	ss $opt | grep "pid=$1" | tr -s ' ' | sed  's/\*/%/g' |
	{
		while IFS= read -r line
		do
           	     echo $line | cut -d ' ' -f 4 > i
               	     check=$(cat i)
#checking if ingress ip= * as format of ss stdout is different in that case
			if [ "$check" = "%" ]
			then
				ingrsip[$count]=$(echo "*")
				ingrsport[$count]=$(echo $line | cut -d ' ' -f 5)
				egrsip[$count]=$(echo "*")
				egrsport[$count]=$(echo $line | cut -d ' ' -f 7)

#checking if firct charecter of ip name is /
#as the ss stdout format is different then
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

		ingrscount=0
		egrscount=0
		count=$((count-1))
		if [ $count != 0 ]
		then
			if [ $x = 1 ]
			then
			echo -e "{\n ingress : [" > netingrs.json
			fi
			while [ $ingrscount -le $count ]
			do
				echo -e "  {\n   ip : \"${ingrsip[$ingrscount]}\" ,\n   port : \"${ingrsport[$ingrscount]}\" ,\n   protocol : \"$prot\"" >> netingrs.json
				if [ $ingrscount = $count ]
				then
					echo "  }" >> netingrs.json
				else
					echo "  }," >> netingrs.json
				fi
				ingrscount=$((ingrscount+1))
			done
			if [ $x = 1 ]
			then
			echo -e " ],\n egress : [" > netegrs.json
			fi
			while [ $egrscount -le $count ]
			do
				echo -e "  {\n   ip : \"${egrsip[$egrscount]}\" ,\n   port : \"${egrsport[$egrscount]}\" ,\n   protocol : \"$prot\"" >> netegrs.json
				if [ $egrscount = $count ]
				then
					echo "  }" >> netegrs.json
				else
					echo "  }," >> netegrs.json
				fi

				egrscount=$((egrscount+1))
			done
			if [ $x = 2 ]
			then
			echo -e " ]\n}" >> netegrs.json
			fi
		elif [ $x = 1 ]
		then
			echo -e "{\n ingress : [" > netingrs.json
			echo -e " ],\n egress : [" > netegrs.json
		else
			echo -e "],\n}" >> netegrs.json
		fi
	}
prot="tcp"
opt="-ntp"
x=$((x+1))
done
}
cat netegrs.json >> netingrs.json
cat netingrs.json > netinfo.json
