#!bash
cd src/

go build -o node ./main.go  ./logger.go

mv node ../node
cd ../