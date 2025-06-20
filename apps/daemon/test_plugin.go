package main

import (
	"fmt"
	"plugin"
	"reflect"

	"github.com/daytonaio/daemon/pkg/toolbox/plugin_loader"
)

func main() {
	// Try to load the plugin
	plug, err := plugin.Open("pkg/toolbox/computeruse_plugin/computeruse.so")
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		return
	}

	// Look up the exported symbol
	sym, err := plug.Lookup(plugin_loader.PluginSymbolName)
	if err != nil {
		fmt.Printf("Failed to find symbol %s: %v\n", plugin_loader.PluginSymbolName, err)
		return
	}

	fmt.Printf("Symbol type: %T\n", sym)
	fmt.Printf("Symbol value: %v\n", sym)

	// Try type assertion
	impl, ok := sym.(plugin_loader.PluginInterface)
	if !ok {
		fmt.Printf("Type assertion failed. Expected PluginInterface, got %T\n", sym)

		// Get more details about the interface
		interfaceType := reflect.TypeOf((*plugin_loader.PluginInterface)(nil)).Elem()
		symbolType := reflect.TypeOf(sym)

		fmt.Printf("Interface type: %v\n", interfaceType)
		fmt.Printf("Symbol type: %v\n", symbolType)

		// Check if the symbol implements the interface methods
		for i := 0; i < interfaceType.NumMethod(); i++ {
			method := interfaceType.Method(i)
			fmt.Printf("Interface method: %s\n", method.Name)

			if _, found := symbolType.MethodByName(method.Name); !found {
				fmt.Printf("  -> Symbol does NOT have method: %s\n", method.Name)
			} else {
				fmt.Printf("  -> Symbol has method: %s\n", method.Name)
			}
		}
		return
	}

	fmt.Printf("Plugin loaded successfully: %T\n", impl)
}
