package main

import (
	"context"
	"encoding/json"
	"fmt"
	argoTriggers "github.com/argoproj/argo-events/sensors/triggers"
	"github.com/kiowy-org/argo-events-jsonpatch-trigger/proto"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
	"sigs.k8s.io/yaml"
)

type JSONPatchTrigger struct {
	K8sClient     kubernetes.Interface
	DynamicClient dynamic.Interface

	namespaceableDynamicClient dynamic.NamespaceableResourceInterface
	proto.TriggerServer
}

func (t *JSONPatchTrigger) FetchResource(ctx context.Context, request *proto.FetchResourceRequest) (*proto.FetchResourceResponse, error) {
	// Extract gvr from the request
	var resource map[string]string
	if err := yaml.Unmarshal(request.Resource, &resource); err != nil {
		return nil, err
	}

	data, err := json.Marshal(resource["target"])
	if err != nil {
		return nil, err
	}
	objData := make(map[string]interface{})
	err = json.Unmarshal(data, &objData)
	obj := unstructured.Unstructured{Object: objData}

	// Fetch the resource from the k8s cluster
	gvr := argoTriggers.GetGroupVersionResource(&obj)
	t.namespaceableDynamicClient = t.DynamicClient.Resource(gvr)
	//since we are patching, we always retrieve a live object
	objName := obj.GetName()
	if objName == "" {
		return nil, fmt.Errorf("resource name is required")
	}

	// for now, we don't support ClusterWide objects
	//todo: add support for cluster wide objects
	objNamespace := obj.GetNamespace()
	if objNamespace == "" {

		return nil, fmt.Errorf("resource namespace is required")
	}
	rObj, err := t.namespaceableDynamicClient.Namespace(objNamespace).Get(ctx, objName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(rObj)
	if err != nil {
		return nil, err
	}
	return &proto.FetchResourceResponse{Resource: body}, nil
}

func (t *JSONPatchTrigger) Execute(ctx context.Context, request *proto.ExecuteRequest) (*proto.ExecuteResponse, error) {
	// Extract object from the request
	var resource map[string]string
	if err := yaml.Unmarshal(request.Resource, &resource); err != nil {
		return nil, err
	}

	//t.namespaceableDynamicClient.Namespace().Patch(ctx, name, types.JSONPatchType, patch, metav1.PatchOptions{})
	return &proto.ExecuteResponse{
		Response: nil,
	}, nil
}

func (t *JSONPatchTrigger) ApplyPolicy(ctx context.Context, request *proto.ApplyPolicyRequest) (*proto.ApplyPolicyResponse, error) {
	//for now, we don't implement any policy so always return success
	return &proto.ApplyPolicyResponse{
		Success: true,
		Message: "success",
	}, nil
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

	proto.RegisterTriggerServer(srv, trigger)

	if err := srv.Serve(lis); err != nil {
		panic(err.Error())
	}
}
