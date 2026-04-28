# How to generate gRPC file
- Login into wsl
```bash
wsl
```
- Running this command
```bash
protoc --go_out=. --go-grpc_out=. pathfolder/pathfile.proto
```