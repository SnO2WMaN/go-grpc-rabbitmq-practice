
## Setup

settings for VSCode

- write `readlink $(which go)` to `go.goroot`
- write `readlink $(which clang-format)` to `clang-format.executable`

## Example

```bash
$ grpcurl -plaintext -d '{ "name": "world" }' localhost:50051 sayhello.Greeter.SayHello
```
