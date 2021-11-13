package util

import (
	"testing"

	"github.com/matryer/is"
)

func TestParseDateTime(t *testing.T) {
	is := is.New(t)

	t.Run("valid dateTime", func(t *testing.T) {
		time, err := ParseDateTime("2020-11-21T16:46:28.2328113")
		is.NoErr(err)
		is.Equal(time.Year(), 2020)
		is.Equal(int(time.Month()), 11)
		is.Equal(time.Day(), 21)
		is.Equal(time.Hour(), 16)
		is.Equal(time.Minute(), 46)
		is.Equal(time.Second(), 28)
	})

	t.Run("invalid dateTime", func(t *testing.T) {
		_, err := ParseDateTime("2020-11-21safasdf")

		is.True(err != nil)
	})
}

func TestParseDate(t *testing.T) {
	is := is.New(t)

	time, err := ParseDate("2020-11-21")
	is.NoErr(err)
	is.Equal(time.Year(), 2020)
	is.Equal(int(time.Month()), 11)
	is.Equal(time.Day(), 21)
	is.Equal(time.Hour(), 0)
	is.Equal(time.Minute(), 0)
	is.Equal(time.Second(), 0)
}

func TestParseInvalidDate(t *testing.T) {
	is := is.New(t)

	_, err := ParseDate("2020-!!!!!")

	is.True(err != nil)
}

func TestFormatDateTime(t *testing.T) {
	is := is.New(t)

	time, _ := ParseDateTime("2020-11-21T16:46:28.2328113")

	formattedTime := FormatDateTime(time)
	is.Equal(formattedTime, "2020-11-21T16:46:28.2328113")
}
