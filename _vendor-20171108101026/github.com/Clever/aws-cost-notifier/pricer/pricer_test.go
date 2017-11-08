package pricer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriceInfo(t *testing.T) {
	assert := assert.New(t)

	created := time.Now().Add(-23*time.Hour - 30*time.Minute)
	c3largePricePerHr := 0.120
	c3largePrice := 24 * c3largePricePerHr // AWS always rounds up to the nearest hour

	pricer, err := FromMap(map[string]float64{"c3.large": c3largePricePerHr})
	assert.NoError(err)

	p, err := pricer.TotalSpent("c3.large", created)
	assert.NoError(err)
	assert.Equal(c3largePrice, p)

	pph, err := pricer.PerHr("c3.large")
	assert.NoError(err)
	assert.Equal(c3largePricePerHr, pph)
}
