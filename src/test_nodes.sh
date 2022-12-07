#!bash

rm "bootstrap_test.txt"

sh ./compile_bootstrap.sh
sh ./compile_node.sh


trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

sh ./run_bootstrap.sh -port 3999 -bootstrap "bootstrap_test.txt" &

sleep 3

sh ./run_bootstrap.sh -port 3998 -bootstrap "bootstrap_test.txt" &

sleep 3

sh ./run_bootstrap.sh -port 3997 -bootstrap "bootstrap_test.txt" &

sleep 3

for i in {5000..5999..20}
do
  sh ./run_node.sh -port $i -bootstrap "bootstrap_test.txt" &
done

sleep 3


for i in {4000..4500..250}
do
  sh ./run_node.sh -port $i -bootstrap "bootstrap_test.txt" &
done

for i in {4500..5000..250}
do
  sh ./run_node.sh -port $i -bootstrap "bootstrap_test.txt" &
done

sleep 3

sh ./run_node.sh -port 6000 -bootstrap "bootstrap_test.txt" &

sleep 3

sh ./run_node.sh -port 6001 -bootstrap "bootstrap_test.txt" &


sleep 20

trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT
