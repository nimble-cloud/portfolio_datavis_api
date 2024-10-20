package api

import (
	"fmt"
	"nimblecloud/portfolio_datavis/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboard(t *testing.T) {
	config.InitEnv()
	config.InitDB()

	rec, c := config.MakeEchoTest("")

	assert.NoError(t, Dashboard(c))

	fmt.Println("dash body -> ", rec.Body.String())

}

func TestRevByState(t *testing.T) {
	config.InitEnv()
	config.InitDB()

	rec, c := config.MakeEchoTest("")
	c.SetPath("/:company/rev/:state")
	c.SetParamNames("state")
	c.SetParamValues("AE")

	assert.NoError(t, RevByState(c))

	fmt.Println("RevByState body -> ", rec.Body.String())
}

func TestGetStates(t *testing.T) {
	config.InitEnv()
	config.InitDB()

	rec, c := config.MakeEchoTest("")

	assert.NoError(t, RevStates(c))

	fmt.Println("RevByState body -> ", rec.Body.String())
}
