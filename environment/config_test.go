package environment_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/goplugin/helmenv/environment"
	"github.com/stretchr/testify/require"
)

func TestChartsFile(t *testing.T) {
	t.Parallel()

	chartsTestFilePath := "./charts-test-file.json"
	pluginImage := "test/plugin/image"
	pluginVersion := "v0.6.4"
	gethImage := "test/geth/image"
	gethVersion := "v0.6.5"

	chartsTestFile, err := os.Create(chartsTestFilePath)
	require.NoError(t, err)
	defer func() { // Cleanup after test
		require.NoError(t, chartsTestFile.Close(), "Error closing test charts file")
		if _, err = os.Stat(chartsTestFilePath); err == nil {
			require.NoError(t, os.Remove(chartsTestFilePath), "Error deleting test charts file")
		}
	}()

	_, err = chartsTestFile.WriteString(fmt.Sprintf(`{
		"geth":{
			"values":{
				 "geth":{
						"image":{
							 "image":"%s",
							 "version":"%s"
						}
				 }
			}
	 },
		"plugin":{
			 "values":{
					"plugin":{
						 "image":{
								"image":"%s",
								"version":"%s"
						 }
					}
			 }
		}
	}`, gethImage, gethVersion, pluginImage, pluginVersion))
	require.NoError(t, err)
	err = chartsTestFile.Sync()
	require.NoError(t, err)

	pluginConfig := environment.NewPluginConfig(map[string]interface{}{}, "")
	err = pluginConfig.Charts.Decode(chartsTestFilePath)
	require.NoError(t, err)
}
