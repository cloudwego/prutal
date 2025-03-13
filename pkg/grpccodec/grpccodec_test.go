package grpccodec

import "testing"

func TestProtoCodec(t *testing.T) {
	type TestStruct struct {
		V string `protobuf:"bytes,1"`
	}

	p0 := &TestStruct{"TestProtoCodec"}
	p1 := &TestStruct{}

	x := &protoCodec{}
	b, err := x.Marshal(p0)
	if err != nil {
		t.Fatal(err)
	}
	if err := x.Unmarshal(b, p1); err != nil {
		t.Fatal(err)
	}
	if p0.V != p1.V {
		t.Fatalf("not equal: %q != %q", p0.V, p1.V)
	}

}
