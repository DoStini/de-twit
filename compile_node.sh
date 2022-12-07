#!bash
cd src/

go build -o node  ./main.go ./utils.go ./dht.go ./logger.go

mv node ../node
cd ../