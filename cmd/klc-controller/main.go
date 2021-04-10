package main

import (
	"flag"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/tivizi/kvm-lvm-csi/endpoint"
	"github.com/tivizi/kvm-lvm-csi/pkg"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(1)
	sock := "unix://tmp/csi-controller.sock"

	listener, _, err := endpoint.Listen(sock)
	if err != nil {
		glog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(pkg.LogGRPC),
	}
	server := grpc.NewServer(opts...)
	csi.RegisterControllerServer(server, &pkg.Driver{})

	glog.Infof("Listening for connections on address: %#v", listener.Addr())
	go server.Serve(listener)
	wg.Wait()
}
