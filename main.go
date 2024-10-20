package main

import (
	"nimblecloud/portfolio_datavis/api"
	"nimblecloud/portfolio_datavis/auth"
	"nimblecloud/portfolio_datavis/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	config.InitEnv()
	config.InitDB()

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${status} ${method} ${uri} ${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	e.POST("/login", auth.Login)
	e.POST("/create-user", auth.CreateUser)

	r := e.Group("/api/v1")
	r.Use(auth.IsAuthed)

	r.GET("/:company/dashboard", api.Dashboard)

	r.GET("/:company/rev/:state", api.RevByState)
	r.GET("/:company/rev/states", api.RevStates)

	r.GET("/:company/reports", api.GetReports)
	r.POST("/:company/report", api.GetReport)
	r.POST("/:company/report-upload", api.UploadReport)
	r.GET("/:company/uploads", api.GetUploadRecords)

	admin := r.Group("/admin")

	admin.POST("/create-user", auth.CreateUser)
	admin.GET("/users", func(c echo.Context) error {
		return c.String(200, "Admin users")
	})

	e.Logger.Fatal(e.Start(":1323"))
}

/*
Functions for visuals are made
company_rev_state()
company_rev() (may not be needed anymore)
avg_order_size()
new_customers()
top_customer()

Jake — Today at 9:57 AM
The below will need to be consumable from the table Company.public.reports table
Functions for reports:
proj_rev_by_tenure
proj_rev_by_tenure2
proj_top_10_percent
proj_rev_breakdown
proj_top_customers
proj_top_customers_new
customer_revenue
*/
