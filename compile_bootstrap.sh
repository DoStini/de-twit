#!bash
cd src/

go build -o bootstrap  ./bootstrap.go ./utils.go ./dht.go ./logger.go

mv bootstrap ../bootstrap
cd ../
