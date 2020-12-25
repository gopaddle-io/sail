FROM dev:latest
	
	USER ria

	

	COPY env_list.log /home/ria/.profile

	CMD  && /bin/bash