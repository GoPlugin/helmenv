package environment

import (
	"path/filepath"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/goplugin/helmenv/tools"
)

const (
	mockServerConfigChartName = "mockserver-config"
	mockServerChartName       = "mockserver"
	pluginChartName        = "plugin"
)

// NewPluginChart returns a default Plugin Helm chart based on a set of override values
func NewPluginChart(index int, values map[string]interface{}) *HelmChart {
	return &HelmChart{Values: values, Index: index}
}

// NewPluginCCIPReorgConfig returns a Plugin environment for the purpose of CCIP testing
func NewPluginCCIPReorgConfig(pluginValues map[string]interface{}, networkIDs []int) *Config {
	return &Config{
		NamespacePrefix: "plugin-ccip",
		Charts: Charts{
			"geth-reorg": {
				Index:       1,
				ReleaseName: "geth-reorg",
				Path:        filepath.Join(tools.ChartsRoot, "geth-reorg"),
				Values: map[string]interface{}{
					"geth": map[string]interface{}{
						"genesis": map[string]interface{}{
							"networkId": networkIDs[0],
						},
					},
				},
			},
			"geth-reorg-2": {
				Index:       1,
				ReleaseName: "geth-reorg-2",
				Path:        filepath.Join(tools.ChartsRoot, "geth-reorg"),
				Values: map[string]interface{}{
					"geth": map[string]interface{}{
						"genesis": map[string]interface{}{
							"networkId": networkIDs[1],
						},
					},
				},
			},
			"plugin": NewPluginChart(3, PluginReplicas(5, pluginValues)),
		},
	}
}

// NewTerraPluginConfig returns a Plugin environment designed for testing with a Terra relay
func NewTerraPluginConfig(pluginValues map[string]interface{}) *Config {
	return &Config{
		NamespacePrefix: "plugin-terra",
		Charts: Charts{
			"localterra": {Index: 1},
			"geth-reorg": {Index: 2},
			"plugin":  NewPluginChart(3, PluginReplicas(2, pluginValues)),
		},
	}
}

// NewPluginReorgConfig returns a Plugin environment designed for simulating re-orgs within testing
func NewPluginReorgConfig(pluginValues map[string]interface{}) *Config {
	return &Config{
		NamespacePrefix: "plugin-reorg",
		Charts: Charts{
			"geth-reorg": {Index: 1},
			"plugin":  NewPluginChart(2, pluginValues),
		},
	}
}

// Organizes passed in values for simulated network charts
type networkChart struct {
	Replicas int
	Values   map[string]interface{}
}

// NewPluginConfig returns a vanilla Plugin environment used for generic functional testing. Geth networks can
// be passed in to launch differently configured simulated geth instances.
func NewPluginConfig(
	pluginValues map[string]interface{},
	optionalNamespacePrefix string,
	networks ...SimulatedNetwork,
) *Config {
	charts := Charts{
		mockServerConfigChartName: {Index: 1},
		mockServerChartName:       {Index: 2},
		pluginChartName:        NewPluginChart(2, pluginValues),
	}

	nameSpacePrefix := loadNetworkCharts(optionalNamespacePrefix, charts, networks)

	return &Config{
		NamespacePrefix: nameSpacePrefix,
		Charts:          charts,
	}
}

// NewPerformancePluginConfig launches an environment with upgraded resources
// Mockserver launches with 1 CPU and 1 GB of RAM
// Plugin DB launches with 1 CPU and 2GB of RAM
// Plugin node launches with 2 CPU and 4GB RAM
func NewPerformancePluginConfig(
	pluginValues map[string]interface{},
	optionalNamespacePrefix string,
	networks ...SimulatedNetwork,
) *Config {
	// Set the plugin value resources to a higher level
	pluginValues["plugin"] = map[string]interface{}{
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "2",
				"memory": "4096Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "2",
				"memory": "4096Mi",
			},
		},
	}
	pluginValues["db"] = map[string]interface{}{
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    "1",
				"memory": "2048Mi",
			},
			"limits": map[string]interface{}{
				"cpu":    "1",
				"memory": "2048Mi",
			},
		},
	}
	mockServerValues := map[string]interface{}{
		"app": map[string]interface{}{
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu":    "1",
					"memory": "1024Mi",
				},
				"limits": map[string]interface{}{
					"cpu":    "1",
					"memory": "1024Mi",
				},
			},
		},
	}
	charts := Charts{
		mockServerConfigChartName: {Index: 1},
		mockServerChartName:       {Index: 2, Values: mockServerValues},
		pluginChartName:        NewPluginChart(2, pluginValues),
	}

	nameSpacePrefix := loadNetworkCharts(optionalNamespacePrefix, charts, networks)

	return &Config{
		NamespacePrefix: nameSpacePrefix,
		Charts:          charts,
	}
}

// loads and properly configures the network charts, and builds a proper namespace config
func loadNetworkCharts(optionalNamespacePrefix string, charts Charts, networks []SimulatedNetwork) string {
	nameSpacePrefix := "plugin"
	if optionalNamespacePrefix != "" {
		nameSpacePrefix = optionalNamespacePrefix
	}

	networkCharts := map[string]*networkChart{}
	for _, networkFunc := range networks {
		chartName, networkValues := networkFunc()
		if networkValues == nil {
			networkValues = map[string]interface{}{}
		}
		// TODO: If multiple networks with the same chart name are present, only use the values from the first one.
		// This means that we can't have mixed network values with the same type
		// (e.g. all geth deployments need to have the same values).
		// Enabling different behavior is a bit of a niche case.
		if _, present := networkCharts[chartName]; !present {
			networkCharts[chartName] = &networkChart{Replicas: 1, Values: networkValues}
		} else {
			if !reflect.DeepEqual(networkValues, networkCharts[chartName].Values) {
				log.Warn().Msg("If trying to launch multiple networks with different underlying values but the same type, " +
					"(e.g. 1 geth performance and 1 geth realistic), that behavior is not currently fully supported. " +
					"Only replicas of the first of that network type will be launched.")
			}
			networkCharts[chartName].Replicas++
		}
	}

	for chartName, networkChart := range networkCharts {
		networkChart.Values["replicas"] = networkChart.Replicas
		charts[chartName] = &HelmChart{Index: 1, Values: networkChart.Values}
	}
	return nameSpacePrefix
}

// SimulatedNetwork is a function that enables launching a simulated network with a returned chart name
// and corresponding values
type SimulatedNetwork func() (string, map[string]interface{})

// DefaultGeth sets up a basic, low-power simulated geth instance. Really just returns empty map to use default values
func DefaultGeth() (string, map[string]interface{}) {
	return "geth", map[string]interface{}{}
}

// PerformanceGeth sets up the simulated geth instance with more power, bigger blocks, and faster mining
func PerformanceGeth() (string, map[string]interface{}) {
	values := map[string]interface{}{}
	values["resources"] = map[string]interface{}{
		"requests": map[string]interface{}{
			"cpu":    "1",
			"memory": "1024Mi",
		},
		"limits": map[string]interface{}{
			"cpu":    "1",
			"memory": "1024Mi",
		},
	}
	values["config_args"] = map[string]interface{}{
		"--dev.period":      "1",
		"--miner.threads":   "1",
		"--miner.gasprice":  "10000000000",
		"--miner.gastarget": "30000000000",
		"--cache":           "4096",
	}
	return "geth", values
}

// RealisticGeth sets up the simulated geth instance to emulate the actual ethereum mainnet as close as possible
func RealisticGeth() (string, map[string]interface{}) {
	values := map[string]interface{}{}
	values["resources"] = map[string]interface{}{
		"requests": map[string]interface{}{
			"cpu":    "1",
			"memory": "1024Mi",
		},
		"limits": map[string]interface{}{
			"cpu":    "1",
			"memory": "1024Mi",
		},
	}
	values["config_args"] = map[string]interface{}{
		"--dev.period":      "14",
		"--miner.threads":   "1",
		"--miner.gasprice":  "10000000000",
		"--miner.gastarget": "15000000000",
		"--cache":           "4096",
	}

	return "geth", values
}

// PluginVersion sets the version of the plugin image to use
func PluginVersion(version string, values map[string]interface{}) map[string]interface{} {
	if values == nil {
		values = map[string]interface{}{}
	}
	values["plugin"] = map[string]interface{}{
		"image": map[string]interface{}{
			"version": version,
		},
	}
	return values
}

// PluginReplicas sets the replica count of plugin nodes to use
func PluginReplicas(count int, values map[string]interface{}) map[string]interface{} {
	if values == nil {
		values = map[string]interface{}{}
	}
	values["replicas"] = count
	return values
}
