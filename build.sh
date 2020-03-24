export PATH=$PATH:$PWD/bin
export GOBIN=$PWD/bin
export RPC_PATH=common

go get -u github.com/golang/protobuf/protoc-gen-go@v1.3.4
go get -u github.com/amsokol/protoc-gen-gotag

# mkdir -p $RPC_PATH
rm -f $RPC_PATH/*.pb.go
protoc -I$RPC_PATH --go_out=plugins=grpc:$RPC_PATH $RPC_PATH/gostfix.proto

protoc -I$RPC_PATH --gotag_out=xxx="bson+\"-\"",output_path=$RPC_PATH:. $RPC_PATH/gostfix.proto

echo "Installing data"
rm -rf data
mkdir data
cp -a main.ini data/
cp -a main.cf data/
cp -a vmailbox.db data/
cp -a web/assets data/
cp -a web/css data/
cp -a web/js data/
cp -a web/templates data/

go build -o $GOBIN/gostfix
