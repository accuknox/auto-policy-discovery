// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	internalinterfaces "github.com/accuknox/auto-policy-discovery/pkg/discoveredpolicy/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// DiscoveredPolicies returns a DiscoveredPolicyInformer.
	DiscoveredPolicies() DiscoveredPolicyInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// DiscoveredPolicies returns a DiscoveredPolicyInformer.
func (v *version) DiscoveredPolicies() DiscoveredPolicyInformer {
	return &discoveredPolicyInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}