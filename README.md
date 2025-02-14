# Prutal

Prutal is a pure Go alternative to [protocol buffers](https://protobuf.dev), it covers most of the functionality offered by Protocol Buffers.

Prutal aims to minimize code generation as much as possible while ensuring serialization and maintaining good performance.

**Since Prutal is NOT yet ready for production use, we are not providing usage documentation at this time, nor do we guarantee backward compatibility of the interface.**

## Features

| Features | Prutal | Protobuf |
| -- | -- | -- |
| Supported Languages | Go | C++, Java, Python, Go, and more |
| Code Generation | ✅ | ✅ |
| Serialization | ✅ without generating code | ✅ |
| Performance | ⭐️⭐️⭐️⭐️⭐️ | ⭐️⭐️⭐️ |
| Extensibility | 😄 Package | 😡 Plugin |
| Compatibility | ✅  (see Protobuf Compatibility) | ✅ |
| gRPC | 🚧 Coming soon | ✅ |
| Non-Pointer Field | ✅  (aka gogoproto.nullable) | ❌ |


## Protobuf Compatibility

* ✅ Works with code generated by the official Protocol Buffer Go
* ✅ Parsing .proto file. syntax: proto2, proto3, edition 2023
* ✅ Protobuf wire format
    - double, float, bool, string, bytes
    - int32, int64, uint32, uint64, sint32, sint64
    - fixed32, fixed64, sfixed64, sfixed64
    - enum
    - repeated, map, oneof
* ✅ Packed / unpack (proto2)
    - PACKED / EXPANDED (repeated field encoding, edition2023)
* ✅ Reserved
* ✅ Unknown Fields
* ⚠️  JSON support: JSON struct tag only
* ⚠️  Code generation: Go only
* ⚠️  Protocol buffers well-known types [link](https://protobuf.dev/reference/protobuf/google.protobuf/)
    - Prutal is able to generate code by reusing pkg [`google.golang.org/protobuf/type`](https://pkg.go.dev/google.golang.org/protobuf/types/known)
    - Features of type like `Any` would be limited.
* ❌ Groups (proto2) [link](https://protobuf.dev/programming-guides/proto2/#groups)
* ❌ Overriding default scalar values (proto2, edition2023) [link](https://protobuf.dev/programming-guides/proto2/#explicit-default)



## Contributing

Contributor guide: [Contributing](CONTRIBUTING.md).

## License

Prutal is licensed under the terms of the Apache license 2.0. See [LICENSE](LICENSE) for more information.
Dependencies used by `prutal` are listed under [licenses](licenses).
