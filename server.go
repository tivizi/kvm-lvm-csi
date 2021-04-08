package main

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/tivizi/kvm-lvm-csi/endpoint"
	"google.golang.org/grpc"
)

func (driver *Driver) Run() error {
	var wg sync.WaitGroup
	wg.Add(1)
	if err := start("unix://tmp/csi.sock", driver, driver, driver); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func start(ep string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) error {
	listener, _, err := endpoint.Listen(ep)
	if err != nil {
		glog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}
	server := grpc.NewServer(opts...)

	if ids != nil {
		csi.RegisterIdentityServer(server, ids)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}

	glog.Infof("Listening for connections on address: %#v", listener.Addr())

	server.Serve(listener)
	return nil
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	pri := glog.Level(3)
	if info.FullMethod == "/csi.v1.Identity/Probe" {
		// This call occurs frequently, therefore it only gets log at level 5.
		pri = 5
	}
	glog.V(pri).Infof("GRPC call: %s", info.FullMethod)

	v5 := glog.V(5)
	if v5 {
		v5.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	}
	resp, err := handler(ctx, req)
	if err != nil {
		// Always log errors. Probably not useful though without the method name?!
		glog.Errorf("GRPC error: %v", err)
	}

	if v5 {
		v5.Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))

		// In JSON format, intentionally logging without stripping secret
		// fields due to below reasons:
		// - It's technically complicated because protosanitizer.StripSecrets does
		//   not construct new objects, it just wraps the existing ones with a custom
		//   String implementation. Therefore a simple json.Marshal(protosanitizer.StripSecrets(resp))
		//   will still include secrets because it reads fields directly
		//   and more complicated code would be needed.
		// - This is indeed for verification in mock e2e tests. though
		//   currently no test which look at secrets, but we might.
		//   so conceptually it seems better to me to include secrets.
		logGRPCJson(info.FullMethod, req, resp, err)
	}

	return resp, err
}

// logGRPCJson logs the called GRPC call details in JSON format
func logGRPCJson(method string, request, reply interface{}, err error) {
	// Log JSON with the request and response for easier parsing
	logMessage := struct {
		Method   string
		Request  interface{}
		Response interface{}
		// Error as string, for backward compatibility.
		// "" on no error.
		Error string
		// Full error dump, to be able to parse out full gRPC error code and message separately in a test.
		FullError error
	}{
		Method:    method,
		Request:   request,
		Response:  reply,
		FullError: err,
	}

	if err != nil {
		logMessage.Error = err.Error()
	}

	msg, err := json.Marshal(logMessage)
	if err != nil {
		logMessage.Error = err.Error()
	}
	glog.V(5).Infof("gRPCCall: %s\n", msg)
}
