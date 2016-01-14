package pricer

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Pricer interface {
	PerHr(instanceType string) (float64, error)
	TotalSpent(instanceType string, created time.Time) (float64, error)
}

type pricer map[string]float64

func FromMap(prices map[string]float64) (Pricer, error) {
	return pricer(prices), nil
}

func FromFile(path string) (Pricer, error) {
	prices := make(pricer)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		instance := line[0]              // first field is name
		priceString := line[len(line)-1] // last field is price/hour
		price, err := strconv.ParseFloat(strings.Trim(priceString, "$ per Hour"), 64)
		if err != nil {
			return nil, err
		}
		prices[instance] = price
	}

	return prices, nil
}

func (p pricer) PerHr(instanceType string) (float64, error) {
	if perHr, ok := p[instanceType]; ok {
		return perHr, nil
	}

	return 0.0, fmt.Errorf("Unknown instance type: %s", instanceType)
}

func (p pricer) TotalSpent(instanceType string, created time.Time) (float64, error) {
	age := time.Since(created)
	if perHr, ok := p[instanceType]; ok {
		// AWS always rounds up to the nearest hour
		return perHr * math.Ceil(age.Hours()), nil
	}

	return 0.0, fmt.Errorf("Unknown instance type: %s", instanceType)
}
