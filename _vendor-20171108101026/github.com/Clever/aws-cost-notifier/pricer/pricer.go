package pricer

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/Clever/pathio"
)

type Pricer interface {
	PerHr(instanceType string) (float64, error)
	TotalSpent(instanceType string, created time.Time) (float64, error)
}

type pricer map[string]float64

func FromMap(prices map[string]float64) (Pricer, error) {
	return pricer(prices), nil
}

func FromS3(path string) (Pricer, error) {
	prices := pricer(map[string]float64{})

	reader, err := pathio.Reader(path)
	if err != nil {
		return prices, err
	}

	if err := json.NewDecoder(reader).Decode(&prices); err != nil {
		return prices, err
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
