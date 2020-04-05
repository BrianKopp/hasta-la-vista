package deregister

import "net/http"

// CloudProvider interface implements the DrainNodeFromLoadBalancer function
type CloudProvider interface {
	DrainNodeFromLoadBalancer(nodeName string, response http.ResponseWriter, request *http.Request) error
}
