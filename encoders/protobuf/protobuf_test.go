package protobuf_test

import (
	"reflect"
	"testing"

	"github.com/nats-io/nats"
	"github.com/nats-io/nats/test"

	"github.com/nats-io/nats/encoders/protobuf"
	pb "github.com/nats-io/nats/encoders/protobuf/testdata"
)

func NewProtoEncodedConn(tl test.TestLogger) *nats.EncodedConn {
	ec, err := nats.NewEncodedConn(test.NewDefaultConnection(tl), protobuf.PROTOBUF_ENCODER)
	if err != nil {
		tl.Fatalf("Failed to create an encoded connection: %v\n", err)
	}
	return ec
}

func TestProtoMarshalStruct(t *testing.T) {
	s := test.RunDefaultServer()
	defer s.Shutdown()

	ec := NewProtoEncodedConn(t)
	defer ec.Close()
	ch := make(chan bool)

	me := &pb.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	ec.Subscribe("protobuf_test", func(p *pb.Person) {
		if !reflect.DeepEqual(p, me) {
			t.Fatal("Did not receive the correct protobuf response")
		}
		ch <- true
	})

	ec.Publish("protobuf_test", me)
	if e := test.Wait(ch); e != nil {
		t.Fatal("Did not receive the message")
	}
}

func BenchmarkProtobufMarshalStruct(b *testing.B) {
	me := &pb.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	encoder := &protobuf.ProtobufEncoder{}
	for n := 0; n < b.N; n++ {
		if _, err := encoder.Encode("protobuf_test", me); err != nil {
			b.Fatal("Couldn't serialize object", err)
		}
	}
}

func BenchmarkPublishProtobufStruct(b *testing.B) {
	// stop benchmark for set-up
	b.StopTimer()

	s := test.RunDefaultServer()
	defer s.Shutdown()

	ec := NewProtoEncodedConn(b)
	defer ec.Close()
	ch := make(chan bool)

	me := &pb.Person{Name: "derek", Age: 22, Address: "140 New Montgomery St"}
	me.Children = make(map[string]*pb.Person)

	me.Children["sam"] = &pb.Person{Name: "sam", Age: 19, Address: "140 New Montgomery St"}
	me.Children["meg"] = &pb.Person{Name: "meg", Age: 17, Address: "140 New Montgomery St"}

	ec.Subscribe("protobuf_test", func(p *pb.Person) {
		if !reflect.DeepEqual(p, me) {
			b.Fatalf("Did not receive the correct protobuf response")
		}
		ch <- true
	})

	// resume benchmark
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		ec.Publish("protobuf_test", me)
		if e := test.Wait(ch); e != nil {
			b.Fatal("Did not receive the message")
		}
	}
}
