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

var nodeId = flag.String("nodeid", "", "node id")

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(1)
	driver, err := pkg.NewDriver(*nodeId)
	if err != nil {
		panic(err)
	}
	sock := "unix://tmp/csi-node.sock"
	listener, _, err := endpoint.Listen(sock)
	if err != nil {
		glog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(pkg.LogGRPC),
	}
	server := grpc.NewServer(opts...)

	csi.RegisterIdentityServer(server, driver)
	csi.RegisterNodeServer(server, driver)

	glog.Infof("Listening for connections on address: %#v", listener.Addr())
	go server.Serve(listener)
	wg.Wait()
}
