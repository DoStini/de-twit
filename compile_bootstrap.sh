#!bash
cd src/

go build -o bootstrap  ./bootstrap.go ./logger.go

mv bootstrap ../bootstrap
cd ../
