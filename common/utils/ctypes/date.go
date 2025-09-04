package ctypes

import (
	"fmt"
	"strings"
	"time"
)

// CustomDate es un wrapper de time.Time que permite parsear fechas en formato "2006-01-02".
type CustomDate time.Time

// UnmarshalJSON parsea la fecha en el formato "2006-01-02".
func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	// Eliminamos las comillas
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*cd = CustomDate(t)
	return nil
}

// MarshalJSON se usa para convertir la fecha al mismo formato al enviar respuestas.
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", time.Time(cd).Format("2006-01-02"))
	return []byte(formatted), nil
}

func (cd CustomDate) After(date CustomDate) bool {
	// Convertir a time.Time y eliminar la parte de hora
	cdDate := time.Date(time.Time(cd).Year(), time.Time(cd).Month(), time.Time(cd).Day(), 0, 0, 0, 0, time.Time(cd).Location())
	dateDate := time.Date(time.Time(date).Year(), time.Time(date).Month(), time.Time(date).Day(), 0, 0, 0, 0, time.Time(date).Location())
	return cdDate.After(dateDate)
}

func (cd CustomDate) Before(date CustomDate) bool {
	// Convertir a time.Time y eliminar la parte de hora
	cdDate := time.Date(time.Time(cd).Year(), time.Time(cd).Month(), time.Time(cd).Day(), 0, 0, 0, 0, time.Time(cd).Location())
	dateDate := time.Date(time.Time(date).Year(), time.Time(date).Month(), time.Time(date).Day(), 0, 0, 0, 0, time.Time(date).Location())
	return cdDate.Before(dateDate)
}

func FromTime(t time.Time) CustomDate {
	return CustomDate(t)
}
