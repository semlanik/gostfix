export GOPATH=$PWD
export PATH=$PATH:$PWD/bin
export GOBIN=$PWD/bin
export RPC_PATH=$PWD/common

go get github.com/golang/protobuf/protoc-gen-go
go install ./src/github.com/golang/protobuf/protoc-gen-go

# mkdir -p $RPC_PATH
rm -f $RPC_PATH/*.pb.go
protoc -I$RPC_PATH --go_out=plugins=grpc:$RPC_PATH $RPC_PATH/gostfix.proto

echo "Installing data"
rm -rf data
mkdir data
cp -a main.ini data/
cp -a main.cf data/
cp -a vmailbox data/
cp -a web/assets data/
cp -a web/css data/
cp -a web/js data/
cp -a web/templates data/

go get -v
go build -o $GOBIN/gostfix
