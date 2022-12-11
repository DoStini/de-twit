# SDLE First Assignment
 
SDLE First Assignment of group T05G16.

Group members:

1. André Moreira (up201904721@edu.fe.up.pt)
2. Margarida Oliveira (up201906784@edu.fe.up.pt)
3. Nuno Alves (up201908250@edu.fe.up.pt)

This is an implementation of a reliable and decentralized social media service using [Golang](https://go.dev/) and [libp2p](https://libp2p.io/).

# User Guide

## Prerequisites and Dependencies

- Download Docker
- Download Golang 1.19

## Compiling

### Linux

- Run `make` script in the `/de-twit-go` directory. It will download all dependencies and create all necessary folders so it requires an internet connection and a few hundred MB of space.
- Run `docker build -t detwit .` in the `/frontend` directory. It will create a docker image running the frontend of the client.

## Running

### Linux

#### Bootstrap

- Run the `./bootstrap -port 3001 -bootstrap bootstraps/bootstraps.txt` executable in the `/de-twit-go` directory with the following arguments: `port`, where `port` is the port where the provider will listen to requests. The bootstrap is where the bootstrap servers will stored and load bootstrap server addresses.

#### Client

- Run the `./node -port 4002 -serve 5002 -bootstrap bootstraps/bootstraps.txt -username sdle` executable in the `/de-twit-go` directory where `port` is the port that will listen other node connections; `serve` is the port on which the http server will listen to allow interaction with the frontend; `bootstrap` where the bootstrap servers will stored and load bootstrap server addresses.

- Run the `./client.sh username address port` script in the `frontend` folder to run the frontend of our service. Username must be the same as the the one provided to node; address is the address of the node at which the http server is listening; port is where we will be able to access our service in the browser as `localhost:port`
    - Example: `./client.sh sdle http://localhost:5002/ 6002`. We will be able to access the service at `http://localhost:6002`


