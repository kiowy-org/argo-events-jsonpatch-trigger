package main

import (
	"fmt"
	"google.golang.org/grpc"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
)

type JSONPatchTrigger struct {
	K8sClient     kubernetes.Interface
	DynamicClient dynamic.Interface
}

func main() {
	port, ok := os.LookupEnv("JPT_PORT")
	if !ok {
		port = "9000"
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// Start the server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err.Error())
	}
	srv := grpc.NewServer()

	trigger := &JSONPatchTrigger{
		K8sClient:     kubernetes.NewForConfigOrDie(config),
		DynamicClient: dynamic.NewForConfigOrDie(config),
	}

	proto.RegisterTriggerServer()

}
