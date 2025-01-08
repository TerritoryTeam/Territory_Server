protoc --go_out=. .\*.proto

protoc.exe --js_out=import_style=commonjs,binary:. .\*.proto