// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"context"
	httpa "net/http"
	"sync"

	"github.com/bborbe/http"
)

type HttpJsonHandler struct {
	ServeHTTPStub        func(context.Context, *httpa.Request) (interface{}, error)
	serveHTTPMutex       sync.RWMutex
	serveHTTPArgsForCall []struct {
		arg1 context.Context
		arg2 *httpa.Request
	}
	serveHTTPReturns struct {
		result1 interface{}
		result2 error
	}
	serveHTTPReturnsOnCall map[int]struct {
		result1 interface{}
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *HttpJsonHandler) ServeHTTP(arg1 context.Context, arg2 *httpa.Request) (interface{}, error) {
	fake.serveHTTPMutex.Lock()
	ret, specificReturn := fake.serveHTTPReturnsOnCall[len(fake.serveHTTPArgsForCall)]
	fake.serveHTTPArgsForCall = append(fake.serveHTTPArgsForCall, struct {
		arg1 context.Context
		arg2 *httpa.Request
	}{arg1, arg2})
	stub := fake.ServeHTTPStub
	fakeReturns := fake.serveHTTPReturns
	fake.recordInvocation("ServeHTTP", []interface{}{arg1, arg2})
	fake.serveHTTPMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *HttpJsonHandler) ServeHTTPCallCount() int {
	fake.serveHTTPMutex.RLock()
	defer fake.serveHTTPMutex.RUnlock()
	return len(fake.serveHTTPArgsForCall)
}

func (fake *HttpJsonHandler) ServeHTTPCalls(stub func(context.Context, *httpa.Request) (interface{}, error)) {
	fake.serveHTTPMutex.Lock()
	defer fake.serveHTTPMutex.Unlock()
	fake.ServeHTTPStub = stub
}

func (fake *HttpJsonHandler) ServeHTTPArgsForCall(i int) (context.Context, *httpa.Request) {
	fake.serveHTTPMutex.RLock()
	defer fake.serveHTTPMutex.RUnlock()
	argsForCall := fake.serveHTTPArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *HttpJsonHandler) ServeHTTPReturns(result1 interface{}, result2 error) {
	fake.serveHTTPMutex.Lock()
	defer fake.serveHTTPMutex.Unlock()
	fake.ServeHTTPStub = nil
	fake.serveHTTPReturns = struct {
		result1 interface{}
		result2 error
	}{result1, result2}
}

func (fake *HttpJsonHandler) ServeHTTPReturnsOnCall(i int, result1 interface{}, result2 error) {
	fake.serveHTTPMutex.Lock()
	defer fake.serveHTTPMutex.Unlock()
	fake.ServeHTTPStub = nil
	if fake.serveHTTPReturnsOnCall == nil {
		fake.serveHTTPReturnsOnCall = make(map[int]struct {
			result1 interface{}
			result2 error
		})
	}
	fake.serveHTTPReturnsOnCall[i] = struct {
		result1 interface{}
		result2 error
	}{result1, result2}
}

func (fake *HttpJsonHandler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *HttpJsonHandler) recordInvocation(key string, args []interface{}) {
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

var _ http.JsonHandler = new(HttpJsonHandler)
