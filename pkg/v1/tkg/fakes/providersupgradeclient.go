// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/providersupgradeclient"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
)

type ProvidersUpgradeClient struct {
	ApplyUpgradeStub        func(*client.ApplyUpgradeOptions) error
	applyUpgradeMutex       sync.RWMutex
	applyUpgradeArgsForCall []struct {
		arg1 *client.ApplyUpgradeOptions
	}
	applyUpgradeReturns struct {
		result1 error
	}
	applyUpgradeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ProvidersUpgradeClient) ApplyUpgrade(arg1 *client.ApplyUpgradeOptions) error {
	fake.applyUpgradeMutex.Lock()
	ret, specificReturn := fake.applyUpgradeReturnsOnCall[len(fake.applyUpgradeArgsForCall)]
	fake.applyUpgradeArgsForCall = append(fake.applyUpgradeArgsForCall, struct {
		arg1 *client.ApplyUpgradeOptions
	}{arg1})
	fake.recordInvocation("ApplyUpgrade", []interface{}{arg1})
	fake.applyUpgradeMutex.Unlock()
	if fake.ApplyUpgradeStub != nil {
		return fake.ApplyUpgradeStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.applyUpgradeReturns
	return fakeReturns.result1
}

func (fake *ProvidersUpgradeClient) ApplyUpgradeCallCount() int {
	fake.applyUpgradeMutex.RLock()
	defer fake.applyUpgradeMutex.RUnlock()
	return len(fake.applyUpgradeArgsForCall)
}

func (fake *ProvidersUpgradeClient) ApplyUpgradeCalls(stub func(*client.ApplyUpgradeOptions) error) {
	fake.applyUpgradeMutex.Lock()
	defer fake.applyUpgradeMutex.Unlock()
	fake.ApplyUpgradeStub = stub
}

func (fake *ProvidersUpgradeClient) ApplyUpgradeArgsForCall(i int) *client.ApplyUpgradeOptions {
	fake.applyUpgradeMutex.RLock()
	defer fake.applyUpgradeMutex.RUnlock()
	argsForCall := fake.applyUpgradeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ProvidersUpgradeClient) ApplyUpgradeReturns(result1 error) {
	fake.applyUpgradeMutex.Lock()
	defer fake.applyUpgradeMutex.Unlock()
	fake.ApplyUpgradeStub = nil
	fake.applyUpgradeReturns = struct {
		result1 error
	}{result1}
}

func (fake *ProvidersUpgradeClient) ApplyUpgradeReturnsOnCall(i int, result1 error) {
	fake.applyUpgradeMutex.Lock()
	defer fake.applyUpgradeMutex.Unlock()
	fake.ApplyUpgradeStub = nil
	if fake.applyUpgradeReturnsOnCall == nil {
		fake.applyUpgradeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.applyUpgradeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ProvidersUpgradeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.applyUpgradeMutex.RLock()
	defer fake.applyUpgradeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ProvidersUpgradeClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ providersupgradeclient.Client = new(ProvidersUpgradeClient)
