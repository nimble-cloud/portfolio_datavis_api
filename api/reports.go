package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"nimblecloud/portfolio_datavis/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

type Report struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Note string `json:"note"`
}

func GetReports(c echo.Context) error {
	rows, err := config.DB.Query(context.Background(), `SELECT name, path, note FROM public.reports`)
	if err != nil {
		return config.HandleErr(err, "failed to query reports")
	}
	defer rows.Close()

	rs := make([]Report, 0)
	for rows.Next() {
		r := Report{}
		err = rows.Scan(&r.Name, &r.Path, &r.Note)
		if err != nil {
			return config.HandleErr(err, "failed to scan reports")
		}

		rs = append(rs, r)
	}

	return c.JSON(http.StatusOK, rs)
}

type ReportBody struct {
	Path string `json:"path"`
}

const (
	Select = "SELECT * FROM "
)

func GetReport(c echo.Context) error {

	p := ReportBody{}
	if err := c.Bind(&p); err != nil {
		return config.HandleErr(err, "could not bind report body")
	}

	rows, err := config.DB.Query(context.Background(), Select+p.Path)
	if err != nil {
		return config.HandleErr(err, "failed to query function "+p.Path)
	}

	data := make([][]string, 0)

	headers := make([]string, 0)
	cols := rows.FieldDescriptions()
	for _, v := range cols {
		headers = append(headers, v.Name)
	}

	colLen := len(headers)
	for rows.Next() {
		d := make([]interface{}, colLen)
		dPointers := make([]interface{}, colLen)

		for i := range d {
			dPointers[i] = &d[i]
		}

		err := rows.Scan(dPointers...)
		if err != nil {
			return config.HandleErr(err, "failed to scan report into value pointers")
		}

		row := make([]string, colLen)
		for i, val := range d {
			if val == nil {
				row[i] = ""
			} else {
				switch v := val.(type) {
				case pgtype.Numeric:
					f, err := val.(pgtype.Numeric).Float64Value()
					if err != nil {
						fmt.Println("no float val")
						fmt.Println(err)
					}

					row[i] = strconv.FormatFloat(f.Float64, 'f', -1, 64)
				case int64:
					row[i] = strconv.FormatInt(v, 10)
				case float64:
					row[i] = strconv.FormatFloat(v, 'f', -1, 64)
				case bool:
					row[i] = strconv.FormatBool(v)
				case []byte:
					row[i] = string(v)
				case string:
					row[i] = v
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}

		data = append(data, row)
	}

	return MakeCsv(c, headers, data)
}

func MakeCsv(c echo.Context, headers []string, data [][]string) error {

	var b bytes.Buffer
	w := csv.NewWriter(&b)
	err := w.Write(headers)

	if err != nil {
		return err
	}

	err = w.WriteAll(data)
	if err != nil {
		return config.HandleErr(err, "error writing csv body")
	}

	return c.Stream(200, "text/csv", &b)
}

func UploadReport(c echo.Context) error {

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return config.HandleErr(err, "failed to open file")
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.Rows("Sheet1")
	if err != nil {
		return config.HandleErr(fmt.Errorf("no data found in the file - Is the sheet titled 'Sheet1'?"), "No data found in the file - Is the sheet titled 'Sheet1'?")
	}
	defer rows.Close()

	return CompanyOrderDetail(rows, c)
}

var headers = []string{"order_id", "customer_id", "account_name", "first_name", "last_name", "email_home", "email_work", "email_web_login", "date_finalized", "total_order_cost", "date_ordered", "order_type", "order_type_name", "addr_shipping_city", "addr_shipping_state", "job_id", "job_name", "is_inv_pull", "job_type", "product_quantity_one", "product_name", "style_id", "product_group", "price_per_unit", "product_quantity_two"}

func CompanyOrderDetail(rows *excelize.Rows, c echo.Context) error {
	fileName := c.FormValue("fileName")
	report := c.FormValue("report")
	notes := c.FormValue("notes")

	// any given report may overlap with existing data. Simply delete all existing data that falls within this report's timeframe
	// column k (Order Date) is the date that matters
	var minDate, maxDate time.Time

	inserts := make([][]any, 0)

	// clean up the data here
	// noticed that the string NULL is found in some fields - will check every value sadly
	// assume all rows could have empty cells ie null
	// finally verify the type is going to be what is expected - dont want to dump out a cryptic error to the customer so be very specific if something isn't right

	rowCount := 1
	for rows.Next() {

		row, err := rows.Columns()
		if err != nil {
			return config.HandleErr(fmt.Errorf("no data found in the file - Is the sheet titled 'Sheet1'?"), "No data found in the file - Is the sheet titled 'Sheet1'?")
		}

		if rowCount == 1 {
			// the first row should be the header and should start with OrderID
			if row[0] != "OrderID" {
				return config.HandleErr(fmt.Errorf("the first row should be the header and should start with OrderID"), "The first row should be the header and should start with OrderID")
			}

			//check the length of the row
			if len(row) != 25 {
				return config.HandleErr(fmt.Errorf("the header row should have 25 columns - Expected: "), "The header row should have 25 columns - Expected: ")
			}
		} else {
			irow := make([]any, 25)

			// find the nulls and verify types
			// add them to the insert row with the correct type
			for ii := 0; ii < len(row); ii++ {

				colNum := ii + 1
				cell := row[ii]

				if cell == "NULL" {
					cell = ""
				}

				// J (10) AND X (25) if currency as a float - verify it is a float
				if colNum == 10 || colNum == 24 {

					var v sql.NullFloat64
					if cell != "" {
						f, err := strconv.ParseFloat(cell, 64)
						if err != nil {
							e := fmt.Errorf("row %d, Column %d - Expected a decimal value but got '%s'", rowCount, colNum, cell)
							return config.HandleErr(e, e.Error())
						}

						v.Float64 = f
						v.Valid = true
					} else {
						v.Valid = false
					}

					irow[ii] = v
					continue // to the next cell
				}

				// I (9) and K (11) if date as a string - verify it is a date and format it to sql date
				if colNum == 9 || colNum == 11 {
					var v sql.NullString
					if cell != "" {
						sqlDate, err := time.Parse("January 2, 2006", cell)
						if err != nil {
							e := fmt.Errorf("row %d, Column %d - Expected a date value but got '%s'", rowCount, colNum, cell)
							return config.HandleErr(e, e.Error())
						}
						v.String = sqlDate.Format("2006-01-02")
						v.Valid = true

						if colNum == 11 {
							if sqlDate.Before(minDate) || minDate.IsZero() {
								minDate = sqlDate
							} else if sqlDate.After(maxDate) || maxDate.IsZero() {
								maxDate = sqlDate
							}
						}

					} else {
						v.Valid = false
					}

					irow[ii] = v
					continue // to the next cell
				}

				// L (12) and T (20) and Y (25) should be integers
				if colNum == 12 || colNum == 20 || colNum == 25 {
					var v sql.NullInt64
					if cell != "" {
						f, err := strconv.Atoi(cell)
						if err != nil {
							e := fmt.Errorf("row %d, Column %d - Expected an integer value but got '%s'", rowCount, colNum, cell)
							return config.HandleErr(e, e.Error())
						}

						v.Int64 = int64(f)
						v.Valid = true
					} else {
						v.Valid = false
					}

					irow[ii] = v
					continue // to the next cell
				}

				// R (18) should be a boolean represented as 1 or 0 or NULL
				if colNum == 18 {
					var v sql.NullBool
					if cell != "" {
						if cell == "1" {
							v.Bool = true
						} else if cell == "0" {
							v.Bool = false
						} else {
							e := fmt.Errorf("row %d, Column %d - Expected a boolean value but got '%s'", rowCount, colNum, cell)
							return config.HandleErr(e, e.Error())
						}
						v.Valid = true

					} else {
						v.Valid = false
					}

					irow[ii] = v
					continue // to the next cell
				}

				// fmt.Println(colNum, cell)
				// everything else should be a string
				irow[ii] = sql.NullString{String: cell, Valid: cell != ""}
			}

			inserts = append(inserts, irow)
		}
		rowCount++
	}

	// delete any existing records that may exists within this time period
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	tx, err := config.DB.Begin(ctx)
	if err != nil {
		return config.HandleErr(err, "failed to create report upload transaction")
	}

	_, err = tx.Exec(ctx, `DELETE FROM public.company_order_details WHERE date_ordered BETWEEN $1 AND $2`, minDate, maxDate)
	if err != nil {
		return config.HandleErr(err, "failed to delete existing data")
	}

	// insert the data
	count, err := tx.CopyFrom(ctx, pgx.Identifier{"public", "company_order_details"}, headers, pgx.CopyFromRows(inserts))
	if err != nil {
		e := fmt.Errorf("error inserting data: %s", err)
		return config.HandleErr(e, e.Error())
	}

	// because the new data is never changing I have modified the revenue by state function to generate the data once and save it to
	// it's own table. This needs to be done every time the data changes! Therefore run it now.
	// It will emply the rev_for_dashboard table and the repopulate it with the revenue data...

	_, err = tx.Exec(ctx, "SELECT public.company_rev_state()")
	if err != nil {
		return config.HandleErr(err, "failed to run rev state function")
	}

	err = MakeUploadRecord(fileName, report, notes, tx)
	if err != nil {
		return config.HandleErr(err, "failed to create upload record")
	}

	err = tx.Commit(ctx)
	if err != nil {
		e := fmt.Errorf("error commiting the data: %s", err)
		return config.HandleErr(e, e.Error())
	}

	return c.String(http.StatusCreated, fmt.Sprint(count))
}

func MakeUploadRecord(fileName, report, notes string, tx pgx.Tx) error {
	_, err := tx.Exec(context.Background(), `INSERT INTO public.uploads(report, file_name, notes) VALUES ($1, $2, $3)`, fileName, report, notes)

	return err
}

type UploadRecs struct {
	Date     string  `json:"date"`
	Report   string  `json:"report"`
	FileName string  `json:"fileName"`
	Notes    *string `json:"notes"`
}

func GetUploadRecords(c echo.Context) error {

	reports := make([]UploadRecs, 0)
	rows, err := config.DB.Query(context.Background(), `SELECT TO_CHAR(date, 'MM/DD/YYYY'), report, file_name, notes FROM public.uploads ORDER BY date DESC LIMIT 50;`)
	if err != nil {
		return config.HandleErr(err, "failed to query uploaded reports")
	}
	defer rows.Close()

	for rows.Next() {
		r := UploadRecs{}
		err = rows.Scan(&r.Date, &r.Report, &r.FileName, &r.Notes)
		if err != nil {
			return config.HandleErr(err, "failed to scan uploaded reports")
		}

		reports = append(reports, r)
	}

	return c.JSON(http.StatusOK, reports)
}
