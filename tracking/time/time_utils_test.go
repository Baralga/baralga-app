package time

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

func TestParseDateTimeForm(t *testing.T) {
	is := is.New(t)

	t.Run("valid dateTime", func(t *testing.T) {
		time, err := ParseDateTimeForm("21.11.2020 16:46")
		is.NoErr(err)
		is.Equal(time.Year(), 2020)
		is.Equal(int(time.Month()), 11)
		is.Equal(time.Day(), 21)
		is.Equal(time.Hour(), 16)
		is.Equal(time.Minute(), 46)
		is.Equal(time.Second(), 0)
	})

	t.Run("invalid dateTime", func(t *testing.T) {
		_, err := ParseDateTimeForm("21.11.2020 safasdf")

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

	formattedTime := FormatDateTime(*time)
	is.Equal(formattedTime, "2020-11-21T16:46:28.2328113")
}

func TestFormatTime(t *testing.T) {
	is := is.New(t)

	time, _ := ParseDateTime("2020-11-21T16:46:28.2328113")

	formattedTime := FormatTime(*time)
	is.Equal(formattedTime, "16:46")
}

func TestFormatDate(t *testing.T) {
	is := is.New(t)

	time, _ := ParseDateTime("2020-11-21T16:46:28.2328113")

	formattedTime := FormatDate(*time)
	is.Equal(formattedTime, "2020-11-21")
}

func TestFormatDateDE(t *testing.T) {
	is := is.New(t)

	time, _ := ParseDateTime("2020-11-21T16:46:28.2328113")

	formattedTime := FormatDateDE(*time)
	is.Equal(formattedTime, "21.11.2020")
}

func TestFormatDateDEShort(t *testing.T) {
	is := is.New(t)

	time, _ := ParseDateTime("2020-11-01T16:46:28.2328113")

	formattedTime := FormatDateDEShort(*time)
	is.Equal(formattedTime, "1.11.")
}

func TestCompleteTimeValue(t *testing.T) {
	is := is.New(t)

	t.Run("only short hours", func(t *testing.T) {
		time := CompleteTimeValue("9")
		is.Equal(time, "09:00")
	})

	t.Run("with backslash", func(t *testing.T) {
		time := CompleteTimeValue("10/12")
		is.Equal(time, "10:12")
	})

	t.Run("only hours", func(t *testing.T) {
		time := CompleteTimeValue("10")
		is.Equal(time, "10:00")
	})

	t.Run("hours with comma ,5", func(t *testing.T) {
		time := CompleteTimeValue("10,5")
		is.Equal(time, "10:30")
	})

	t.Run("hours with comma ,75", func(t *testing.T) {
		time := CompleteTimeValue("10,75")
		is.Equal(time, "10:45")
	})

	t.Run("no time", func(t *testing.T) {
		time := CompleteTimeValue("xxx")
		is.Equal(time, "xxx")
	})
}

func TestQuarter(t *testing.T) {
	is := is.New(t)

	time, _ := ParseDateTime("2020-11-01T16:46:28.2328113")

	quarter := Quarter(*time)
	is.Equal(quarter, 4)
}
