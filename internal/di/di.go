package di

import (
	"fmt"
	"reflect"
)

type DIContext struct {
	singletonInstances map[string]any
	singletonFactories map[string]any
	factories          map[string]any

	interfaceSingletonInstances map[string]any
	interfaceSingletonFactories map[string]any
	interfaceFactories          map[string]any
}

func NewContext() *DIContext {
	ctx := new(DIContext)
	ctx.singletonInstances = make(map[string]any)
	ctx.singletonFactories = make(map[string]any)
	ctx.factories = make(map[string]any)

	ctx.interfaceSingletonInstances = make(map[string]any)
	ctx.interfaceSingletonFactories = make(map[string]any)
	ctx.interfaceFactories = make(map[string]any)
	return ctx
}

func AddSingleton[T any](provider *DIContext, factory func(*DIContext) *T) {
	tName := getTypeName[T]()
	provider.singletonFactories[tName] = factory
}

func AddInterfaceSingleton[T any](provider *DIContext, factory func(*DIContext) T) {
	tName := getTypeName[T]()
	provider.interfaceSingletonFactories[tName] = factory
}

func AddFactory[T any](provider *DIContext, factory func(*DIContext) *T) {
	tName := getTypeName[T]()
	provider.factories[tName] = factory
}

func AddInterfaceFactory[T any](provider *DIContext, factory func(*DIContext) T) {
	tName := getTypeName[T]()
	provider.interfaceFactories[tName] = factory
}

func GetService[T any](provider *DIContext) *T {
	tName := getTypeName[T]()

	tGeneric, found := provider.singletonInstances[tName]
	if found {
		t, ok := tGeneric.(*T)

		if ok {
			return t
		}
	}

	serviceFactory, found := provider.singletonFactories[tName]
	if found {
		tFactory, ok := serviceFactory.(func(*DIContext) *T)

		if ok {
			t := tFactory(provider)
			provider.singletonInstances[tName] = t
			return t
		}
	}

	serviceFactory, found = provider.factories[tName]
	if found {
		tFactory, ok := serviceFactory.(func(*DIContext) *T)

		if ok {
			return tFactory(provider)
		}
	}

	var panicMsg string
	if !found {
		panicMsg = fmt.Sprintf("Service %s singleton or factory not found", tName)
	} else {
		panicMsg = fmt.Sprintf("Failed to convert service %s", tName)
	}
	panic(panicMsg)
}

func GetInterfaceService[T any](provider *DIContext) T {
	tName := getTypeName[T]()

	tGeneric, found := provider.interfaceSingletonInstances[tName]
	if found {
		t, ok := tGeneric.(T)

		if ok {
			return t
		}
	}

	serviceFactory, found := provider.interfaceSingletonFactories[tName]
	if found {
		tFactory, ok := serviceFactory.(func(*DIContext) T)

		if ok {
			t := tFactory(provider)
			provider.interfaceSingletonInstances[tName] = t
			return t
		}
	}

	serviceFactory, found = provider.interfaceFactories[tName]
	if found {
		tFactory, ok := serviceFactory.(func(*DIContext) T)

		if ok {
			return tFactory(provider)
		}
	}

	var panicMsg string
	if !found {
		panicMsg = fmt.Sprintf("Interface %s singleton or factory not found", tName)
	} else {
		panicMsg = fmt.Sprintf("Interface to convert service %s", tName)
	}
	panic(panicMsg)
}

func getTypeName[T any]() string {
	tType := reflect.TypeOf([0]T{}).Elem()
	return tType.PkgPath() + "/" + tType.Name()
}
