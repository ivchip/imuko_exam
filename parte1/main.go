package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// Trade clients transations
type trade struct {
	ClientID int     `json:"clientId"`
	Phone    int     `json:"phone"`
	Nombre   string  `json:"nombre"`
	Compro   bool    `json:"compro"`
	Tdc      string  `json:"tdc"`
	Monto    float64 `json:"monto"`
	Date     string  `json:"date"`
}

// Statistics transations statistics
type statistics struct {
	Total         float64            `json:"total"`
	ComprasPorTDC map[string]float64 `json:"comprasPorTDC"`
	NoCompraron   int                `json:"noCompraron"`
	CompraMasAlta float64            `json:"compraMasAlta"`
}

const baseURL = "https://apirecruit-gjvkhl2c6a-uc.a.run.app/compras/%s"
const shortForm = "2006-01-02"

var wg sync.WaitGroup

// getApiCruit Consume API Rest
func getAPICruit(date string, c chan<- []trade) {
	defer wg.Done()
	var trades []trade

	url := fmt.Sprintf(baseURL, date)
	response, err := http.Get(url)
	if err != nil {
		c <- nil
	}

	err = json.NewDecoder(response.Body).Decode(&trades)
	if err != nil {
		c <- nil
	}
	c <- trades
}

func getValidStringMonth(month time.Month) string {
	var numValid string
	numInt, _ := strconv.Atoi(fmt.Sprintf("%d", month))
	if numInt < 10 {
		numValid = fmt.Sprintf("0%d", numInt)
	} else {
		numValid = strconv.Itoa(numInt)
	}
	return numValid
}

func getValidStringDay(num int) string {
	var numValid string
	if num < 10 {
		numValid = fmt.Sprintf("0%d", num)
	} else {
		numValid = strconv.Itoa(num)
	}
	return numValid
}

func getArrayDates(startDate time.Time, days int, dates *[]string) {
	for i := 0; i < days; i++ {
		t1 := startDate.AddDate(0, 0, i)
		stringT1 := fmt.Sprintf("%d-%s-%s", t1.Year(), getValidStringMonth(t1.Month()), getValidStringDay(t1.Day()))
		*dates = append(*dates, stringT1)
	}
}

func getStats(trades []trade, stats *statistics) {
	var valTemp float64
	m := make(map[string]float64)
	for _, t := range trades {
		stats.Total = stats.Total + t.Monto
		value, ok := m[t.Tdc]
		if ok {
			m[t.Tdc] = value + t.Monto
		} else {
			m[t.Tdc] = t.Monto
		}
		if !t.Compro {
			stats.NoCompraron++
		}
		if t.Monto > valTemp {
			valTemp = t.Monto
		}
	}
	stats.CompraMasAlta = valTemp
	delete(m, "")
	stats.ComprasPorTDC = m
}

// Handler resumen
func summaryFunc(ctx echo.Context) error {
	paramDate := ctx.Param("date")
	valueDate, err := time.Parse(shortForm, paramDate)

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	paramDays, err := strconv.Atoi(ctx.QueryParam("dias"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	var trades []trade
	dates := []string{}
	getArrayDates(valueDate, paramDays, &dates)
	c := make(chan []trade, len(dates))

	for _, date := range dates {
		wg.Add(1)
		go getAPICruit(date, c)
	}
	wg.Wait()
	close(c)
	var stats statistics

	for item := range c {
		trades = append(trades, item...)
	}
	getStats(trades, &stats)
	return ctx.JSON(http.StatusOK, stats)
}

func main() {
	// Echo instance
	e := echo.New()
	// Routes
	e.GET("/resumen/:date", summaryFunc)
	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
