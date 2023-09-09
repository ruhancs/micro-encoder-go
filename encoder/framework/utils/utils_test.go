package utils_test

import (
	"encoder/framework/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	json := `{
		"id": "asdgauy34873e",
		"file_path": "convite.mp4",
		"status": "pending"
	}`

	err := utils.IsJson(json)
	require.Nil(t,err)

	json = `MSG`
	err = utils.IsJson(json)
	require.Error(t,err)
}