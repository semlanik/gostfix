export GOBIN=$(go env GOPATH)/bin
export PATH=$PATH:$GOBIN
export RPC_PATH=common

go env
go install google.golang.org/protobuf/compiler/protogen
go install github.com/amsokol/protoc-gen-gotag

# mkdir -p $RPC_PATH
rm -f $RPC_PATH/*.pb.go
protoc -I$RPC_PATH --go_out=plugins=grpc:$PWD $RPC_PATH/gostfix.proto

protoc -I$RPC_PATH --gotag_out=xxx="bson+\"-\"",output_path=$RPC_PATH:. $RPC_PATH/gostfix.proto

#echo "Installing data"
#rm -rf data
#mkdir data
#cp -a main.ini data/
#cp -a main.cf data/
#cp -a vmailbox.db data/
cp -a web/assets data/
cp -a web/css data/
cp -a web/js data/
cp -a web/templates data/

go build -o bin/gostfix
go mod tidy
