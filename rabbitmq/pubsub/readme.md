### Simple publish subscribe example

uses default credentials and port

```sh
# listen to messages
go run consumer.go

# send two messages
go run publisher.go
go run publisher.go
```