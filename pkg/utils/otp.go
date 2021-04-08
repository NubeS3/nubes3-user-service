package utils

import (
	"github.com/thanhpk/randstr"
	"strings"
	"time"
)

type Otp struct {
	Val     string    `json:"otp"`
	Expired time.Time `json:"expired"`
}

func GenerateOtp() *Otp {
	otp := strings.ToUpper(randstr.Hex(4))
	exp := time.Now().Add(time.Minute * 15)
	return &Otp{
		Val:     otp,
		Expired: exp,
	}
}

func SendOtp(fullname, username, email, otp string, exp time.Time) {

}
