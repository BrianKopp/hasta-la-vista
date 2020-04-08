package deregister

// CloudProvider interface implements the DrainNodeFromLoadBalancer function
type CloudProvider interface {
	DrainNodeFromLoadBalancer(nodeName string) error
}
