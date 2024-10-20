package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nimblecloud/portfolio_datavis/config"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type TopTen struct {
	Account string  `json:"account"`
	Revenue float64 `json:"revenue"`
}

type Dash struct {
	AvgOrderSize float64  `json:"avgOrderSize"`
	TopCustomer  string   `json:"topCustomer"`
	NewCustomers int      `json:"newCustomers"`
	TopTen       []TopTen `json:"topTen"`
}

func Dashboard(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	d := Dash{}

	batch := &pgx.Batch{}
	batch.Queue("SELECT public.avg_order_size()")
	batch.Queue("SELECT public.top_customer()")
	batch.Queue("SELECT public.new_customers()")
	batch.Queue("SELECT * FROM public.top_10_customer()") //account_name2, revenue

	br := config.DB.SendBatch(ctx, batch)
	defer func() {
		err := br.Close()
		if err != nil {
			fmt.Println("error closing batch for dashboard!", err)
			// return config.HandleErr(err, "failed to close the dashboard batch")
		}
	}()

	// average order size
	rows, errOne := br.Query()
	if errOne != nil {
		return config.HandleErr(errOne, "failed to query avg_order_size()")
	}

	for rows.Next() {
		errOne = rows.Scan(&d.AvgOrderSize)
		if errOne != nil {
			return config.HandleErr(errOne, "failed to scan avg_order_size()")
		}
	}

	// top customer
	rows, errTwo := br.Query()
	if errTwo != nil {
		fmt.Println("the error is certainly from here")
		return config.HandleErr(errTwo, "failed to query top_customer() foo")
	}

	for rows.Next() {
		errTwo = rows.Scan(&d.TopCustomer)
		if errTwo != nil {
			return config.HandleErr(errTwo, "failed to scan top_customer()")
		}
	}

	// new customers
	rows, errThree := br.Query()
	if errThree != nil {
		return config.HandleErr(errThree, "failed to query new_customers()")
	}

	for rows.Next() {
		errThree = rows.Scan(&d.NewCustomers)
		if errThree != nil {
			return config.HandleErr(errThree, "failed to scan new_customers()")
		}
	}

	tt := make([]TopTen, 10)
	// top 10 customers
	rows, errFour := br.Query()
	if errFour != nil {
		return config.HandleErr(errFour, "failed to query top_10_customer()")
	}

	count := 0
	for rows.Next() {
		t := TopTen{}
		errFour = rows.Scan(&t.Account, &t.Revenue)
		if errFour != nil {
			return config.HandleErr(errFour, "failed to scan top_10_customer()")
		}

		tt[count] = t
		count++
	}

	d.TopTen = tt
	// fmt.Println("d>>", d)

	return c.JSON(http.StatusOK, d)
}

type RevState struct {
	Month int `json:"month"`
	// State   string  `json:"state"`
	Revenue  float64 `json:"revenue"`
	SortType string  `json:"sortType"`
}

func RevByState(c echo.Context) error {
	state := c.Param("state")

	rev := make([]RevState, 0)
	rows, err := config.DB.Query(context.Background(), "SELECT order_month, revenue, type FROM public.rev_for_dashboard WHERE state = $1 ORDER BY order_month ASC", state)
	if err != nil {
		return config.HandleErr(err, "failed to query rev_for_dashboard")
	}
	defer rows.Close()

	for rows.Next() {
		r := RevState{}
		err := rows.Scan(&r.Month, &r.Revenue, &r.SortType)
		if err != nil {
			return config.HandleErr(err, "failed to scan rev_for_dashboard")
		}

		rev = append(rev, r)
	}

	return c.JSON(http.StatusOK, rev)
}

func RevStates(c echo.Context) error {

	states := make([]string, 0)
	rows, err := config.DB.Query(context.Background(), "SELECT DISTINCT state FROM public.rev_for_dashboard ORDER BY state ASC")
	if err != nil {
		return config.HandleErr(err, "failed to query states for revenue dash")
	}

	for rows.Next() {
		s := ""
		err = rows.Scan(&s)
		if err != nil {
			return config.HandleErr(err, "failed to scan states for revenue dash")
		}

		states = append(states, s)
	}

	return c.JSON(http.StatusOK, states)
}
