package api

import (
	"fmt"
	"nimblecloud/portfolio_datavis/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReports(t *testing.T) {
	config.InitEnv()
	config.InitDB()

	rec, c := config.MakeEchoTest("")

	assert.NoError(t, GetReports(c))

	fmt.Println("reports body -> ", rec.Body.String())
}

func TestGetReport(t *testing.T) {
	config.InitEnv()
	config.InitDB()

	rec, c := config.MakeEchoTest(`{"path": "public.reports;"}`)
	assert.NoError(t, GetReport(c))

	fmt.Println("reports body -> ", rec.Body.String())
}
