// Code generated by prutalgen. DO NOT EDIT.
// prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative ./oneof.proto

package oneof

type TestOneofMessage struct {
	// Types that are assignable to OneOfField1:
	//
	//	*TestOneofMessage_Field1
	//	*TestOneofMessage_Field2
	OneOfField1 isTestOneofMessage_OneOfField1 `protobuf_oneof:"one_of_field1"`
	// Types that are assignable to OneOfField2:
	//
	//	*TestOneofMessage_Field3
	//	*TestOneofMessage_Field4
	OneOfField2 isTestOneofMessage_OneOfField2 `protobuf_oneof:"one_of_field2"`
}

func (x *TestOneofMessage) Reset() { *x = TestOneofMessage{} }

// XXX_OneofWrappers is for the internal use of the prutal package.
func (*TestOneofMessage) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*TestOneofMessage_Field1)(nil),
		(*TestOneofMessage_Field2)(nil),
		(*TestOneofMessage_Field3)(nil),
		(*TestOneofMessage_Field4)(nil),
	}
}

type isTestOneofMessage_OneOfField1 interface {
	isTestOneofMessage_OneOfField1()
}

type TestOneofMessage_Field1 struct {
	Field1 bool `protobuf:"varint,1,opt,name=field1" json:"field1,omitempty"`
}

func (*TestOneofMessage_Field1) isTestOneofMessage_OneOfField1() {}

type TestOneofMessage_Field2 struct {
	Field2 int64 `protobuf:"varint,2,opt,name=field2" json:"field2,omitempty"`
}

func (*TestOneofMessage_Field2) isTestOneofMessage_OneOfField1() {}

type isTestOneofMessage_OneOfField2 interface {
	isTestOneofMessage_OneOfField2()
}

type TestOneofMessage_Field3 struct {
	Field3 int32 `protobuf:"varint,3,opt,name=field3" json:"field3,omitempty"`
}

func (*TestOneofMessage_Field3) isTestOneofMessage_OneOfField2() {}

type TestOneofMessage_Field4 struct {
	Field4 *TestOneofMessage `protobuf:"bytes,4,opt,name=field4" json:"field4,omitempty"`
}

func (*TestOneofMessage_Field4) isTestOneofMessage_OneOfField2() {}
