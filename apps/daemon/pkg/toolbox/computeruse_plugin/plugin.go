package main

import "github.com/daytonaio/daemon/pkg/toolbox/plugin_loader"

// ComputerUsePlugin is the exported plugin symbol
var ComputerUsePlugin plugin_loader.PluginInterface = &ComputerUse{}
