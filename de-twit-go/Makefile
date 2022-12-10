all: bootstrap node

bootstrap:
	go build -o bootstrap ./bootstrap.go ./logger.go

node:
	go build -o node ./main.go  ./logger.go

cleanall: clean_bootstraps clean_logs clean_storage clean

clean:
	rm -f bootstrap node

clean_bootstraps:
	rm -f bootstraps/*.txt

clean_logs:
	rm -f logs/*.log

clean_storage:
	rm -rf storage/*
