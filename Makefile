include .env
	
DEV:
	@echo "ENVIRONMENT=DEV" > .env
	@sh sh/dev.sh
	rm -rf go.mod
	rm -rf go.sum
	rm -rf Template
	go mod init Template
	go mod tidy
	go build -o Template
	nohup ./Template
	@echo "SUCCESSFULLY RUN IN PORT ${PORT}"
	
SIT:
	@echo "ENVIRONMENT=SIT" > .env
	@sh sh/sit.sh
	rm -rf go.mod
	rm -rf go.sum
	rm -rf Template
	go mod init Template
	go mod tidy
	go build -o Template
	nohup ./Template
	@echo "SUCCESSFULLY RUN IN PORT ${PORT}"
	
UAT:
	@echo "ENVIRONMENT=UAT" > .env
	@sh sh/uat.sh
	rm -rf go.mod
	rm -rf go.sum
	rm -rf Template
	go mod init Template
	go mod tidy
	go build -o Template
	nohup ./Template
	@echo "SUCCESSFULLY RUN IN PORT ${PORT}"

PROD:
	@echo "ENVIRONMENT=PROD" > .env
	@sh sh/prod.sh
	rm -rf go.mod
	rm -rf go.sum
	rm -rf Template
	go mod init Template
	go mod tidy
	go build -o Template
	nohup ./Template
	@echo "SUCCESSFULLY RUN IN PORT ${PORT}"
	
KILLS: 
	@lsof -n -i4TCP:${PORT} | grep LISTEN | awk '{ print $2 }'| sh kill.sh | echo "SUCCESSFULLY KILLED PROCESS ON PORT ${PORT}"

CHECK: 
	@lsof -i :${PORT}

LOG:
	sh monitor.sh

   
	
