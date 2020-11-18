os_name=$(grep "^NAME" /etc/os-release | sed  's/^[^\"]*"//; s/[\"].*$//; s/\ //; s/[A-Z].*/\L&/')
if [ "$os_name" == "ubuntu" ]
then
	apt update -y
	cat packages.log | apt install -y
elif [ "$os_name" == "archlinux" ]
then
	pacman -Syy
	while IFS= read -r line
	do
		pacman -Syy --noconfirm
		result="$(pacman -Q $line 2>/dev/null)"
		[ ! -z "$result" ] || pacman -S $line --noconfirm
	done < packages.log
elif [ "$os_name" == "centos" ]
then
	while IFS= read -r line
	do
		yum -y install $line
	done < packages.log

fi
