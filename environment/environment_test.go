package environment_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"github.com/goplugin/helmenv/environment"
	"github.com/goplugin/helmenv/tools"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func teardown(t *testing.T, e *environment.Environment) {
	err := e.Teardown()
	require.NoError(t, err)
}

func TestCanDeployAll(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       2, // Deliberate unordered keys to test the OrderedKeys function in Charts
	})
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "plugin",
		Path:        filepath.Join(tools.ChartsRoot, "plugin"),
		Index:       4, // Deliberate unordered keys to test the OrderedKeys function in Charts
	})
	require.NoError(t, err)

	err = e.DeployAll()
	require.NoError(t, err)
	err = e.ConnectAll()
	require.NoError(t, err)

	require.NotEmpty(t, e.Config.Charts["geth"].ChartConnections["geth_0_geth-network"].RemotePorts["ws-rpc"])
	require.NotEmpty(t, e.Config.Charts["geth"].ChartConnections["geth_0_geth-network"].LocalPorts["ws-rpc"])

	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_node"].RemotePorts["access"])
	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_node"].LocalPorts["access"])
	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_plugin-db"].RemotePorts["postgres"])
	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_plugin-db"].LocalPorts["postgres"])
}

func TestMultipleChartsSeparate(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       1,
	})
	require.NoError(t, err)
	err = e.Deploy("geth")
	require.NoError(t, err)
	err = e.Connect("geth")
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "plugin",
		Path:        filepath.Join(tools.ChartsRoot, "plugin"),
		Index:       2,
	})
	require.NoError(t, err)
	err = e.Deploy("plugin")
	require.NoError(t, err)
	err = e.Connect("plugin")
	require.NoError(t, err)

	require.NotEmpty(t, e.Config.Charts["geth"].ChartConnections["geth_0_geth-network"].RemotePorts["ws-rpc"])
	require.NotEmpty(t, e.Config.Charts["geth"].ChartConnections["geth_0_geth-network"].LocalPorts["ws-rpc"])

	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_node"].RemotePorts["access"])
	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_node"].LocalPorts["access"])
	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_plugin-db"].RemotePorts["postgres"])
	require.NotEmpty(t, e.Config.Charts["plugin"].ChartConnections["plugin-node_0_plugin-db"].LocalPorts["postgres"])
}

func TestDeployRepositoryChart(t *testing.T) {
	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "nginx",
		URL:         "https://charts.bitnami.com/bitnami/nginx-9.5.13.tgz",
		Index:       1,
	})
	require.NoError(t, err)

	err = e.Deploy("nginx")
	require.NoError(t, err)
}

func TestParallelDeployments(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       1,
	})
	require.NoError(t, err)
	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "plugin-1",
		Path:        filepath.Join(tools.ChartsRoot, "plugin"),
		Index:       2,
	})
	require.NoError(t, err)
	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "plugin-2",
		Path:        filepath.Join(tools.ChartsRoot, "plugin"),
		Index:       2,
	})
	require.NoError(t, err)
	err = e.DeployAll()
	require.NoError(t, err)

	require.NotEmpty(t, e.Config.Charts["plugin-1"].ChartConnections["plugin-1-node_0_node"].RemotePorts["access"])
	require.NotEmpty(t, e.Config.Charts["plugin-2"].ChartConnections["plugin-2-node_0_node"].RemotePorts["access"])
}

func TestExecuteInPod(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       1,
	})
	require.NoError(t, err)
	err = e.Deploy("geth")
	require.NoError(t, err)
	err = e.Connect("geth")
	require.NoError(t, err)

	err = e.Charts.ExecuteInPod("geth", "geth", 0, "geth-network", []string{"ls", "-a"})

	require.NoError(t, err)
}

func TestUpgrade(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       1,
	})
	require.NoError(t, err)
	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "plugin",
		Path:        filepath.Join(tools.ChartsRoot, "plugin"),
		Index:       2,
	})
	require.NoError(t, err)
	err = e.DeployAll()
	require.NoError(t, err)
	err = e.ConnectAll()
	require.NoError(t, err)

	urls, err := e.Charts.Connections("plugin").LocalURLsByPort("access", environment.HTTP)
	require.NoError(t, err)
	require.Len(t, urls, 1)
	e.Disconnect()

	chart, err := e.Charts.Get("plugin")
	require.NoError(t, err)
	chart.Values = environment.PluginReplicas(2, nil)
	err = chart.Upgrade()
	require.NoError(t, err)

	err = e.ConnectAll()
	require.NoError(t, err)

	urls, err = e.Charts.Connections("plugin").LocalURLsByPort("access", environment.HTTP)
	require.NoError(t, err)
	require.Len(t, urls, 2)

	require.NoError(t, err)
}

func TestBeforeAndAfterHook(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	var before, after string
	err = e.AddChart(&environment.HelmChart{
		BeforeHook: func(_ *environment.Environment) error {
			before = "value"
			return nil
		},
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       1,
		AfterHook: func(_ *environment.Environment) error {
			after = "value"
			return nil
		},
	})
	require.NoError(t, err)
	err = e.Deploy("geth")
	require.NoError(t, err)
	err = e.Connect("geth")
	require.NoError(t, err)

	require.NotEmpty(t, before)
	require.NotEmpty(t, after)
}

func TestAutoConnect(t *testing.T) {
	t.Parallel()

	envName := fmt.Sprintf("test-env-%s", uuid.NewV4().String())
	e, err := environment.NewEnvironment(&environment.Config{})
	defer teardown(t, e)
	require.NoError(t, err)
	err = e.Init(envName)
	require.NoError(t, err)

	err = e.AddChart(&environment.HelmChart{
		ReleaseName: "geth",
		Path:        filepath.Join(tools.ChartsRoot, "geth"),
		Index:       1,
		AutoConnect: true,
		AfterHook: func(_ *environment.Environment) error {
			require.NotEmpty(t, e.Config.Charts["geth"].ChartConnections["geth_0_geth-network"].LocalPorts["ws-rpc"])
			return nil
		},
	})
	require.NoError(t, err)
	err = e.Deploy("geth")
	require.NoError(t, err)
}

func TestCanConnectProgrammatically(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestCanConnectCLI(t *testing.T) {
	t.Parallel()
	// TODO
}
