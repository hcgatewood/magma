package main

import (
	"flag"

	"magma/fbinternal/cloud/go/services/download/servicers"
	"magma/orc8r/cloud/go/plugin"
	"magma/orc8r/cloud/go/services/service_registry"
)

func main() {
	flag.Parse()
	plugin.LoadAllPluginsFatalOnError(&plugin.DefaultOrchestratorPluginLoader{})
	service_registry.MustPopulateServices()

	servicers.Run()
}
