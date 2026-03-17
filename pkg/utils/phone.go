package utils

import (
	"fmt"
	"math/rand"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nyaruka/phonenumbers"
)

func GeneratePhone(countryCode string) string {
	if countryCode == "RU" {
		return fmt.Sprintf("+79%09d", rand.Intn(1_000_000_000))
	}
	num, _ := phonenumbers.Parse(gofakeit.Phone(), countryCode)
	return phonenumbers.Format(num, phonenumbers.E164)
}
