export GOPATH=$PWD
export PATH=$PATH:$PWD/bin
export GOBIN=$PWD/bin
export RPC_PATH=$PWD/common

go get github.com/golang/protobuf/protoc-gen-go
go install ./src/github.com/golang/protobuf/protoc-gen-go

# mkdir -p $RPC_PATH
rm -f $RPC_PATH/*.pb.go
protoc -I$RPC_PATH --go_out=plugins=grpc:$RPC_PATH $RPC_PATH/gostfix.proto

go get -v
go build -o $GOBIN/gostfix

