package placeholder

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/bxcodec/faker/v4"
)

var randomGenerators = map[string]func(args map[string]string) string{
	// === Identificadores / seguridad ===
	"random.UUID": func(args map[string]string) string { // uuid con guiones
		return faker.UUIDHyphenated()
	},
	"random.UUIDDigit": func(args map[string]string) string { // uuid sin guiones
		return faker.UUIDDigit()
	},
	"random.JWT": func(args map[string]string) string {
		return faker.Jwt()
	},

	// === Persona / nombres ===
	"random.Name":      func(args map[string]string) string { return faker.Name() },
	"random.FirstName": func(args map[string]string) string { return faker.FirstName() },
	"random.LastName":  func(args map[string]string) string { return faker.LastName() },

	// === Contacto ===
	"random.Email":     func(args map[string]string) string { return faker.Email() },
	"random.Phone":     func(args map[string]string) string { return faker.Phonenumber() },     // formato genérico
	"random.E164Phone": func(args map[string]string) string { return faker.E164PhoneNumber() }, // +NN...

	// === Internet / red ===
	"random.Username":   func(args map[string]string) string { return faker.Username() },
	"random.URL":        func(args map[string]string) string { return faker.URL() },
	"random.DomainName": func(args map[string]string) string { return faker.DomainName() },
	"random.IPv4":       func(args map[string]string) string { return faker.IPv4() },
	"random.IPv6":       func(args map[string]string) string { return faker.IPv6() },
	"random.MacAddress": func(args map[string]string) string { return faker.MacAddress() },

	// === Texto ===
	"random.Word":      func(args map[string]string) string { return faker.Word() },
	"random.Sentence":  func(args map[string]string) string { return faker.Sentence() },
	"random.Paragraph": func(args map[string]string) string { return faker.Paragraph() },

	// === Password (longitud opcional) ===
	"random.Password": func(args map[string]string) string {
		lengthStr := getArgOr(args, "length", "12")
		n, err := strconv.Atoi(lengthStr)
		if err != nil || n < 4 {
			n = 12
		}
		p := faker.Password()
		if len(p) >= n {
			return p[:n]
		}
		// si faker.Password() salió más corto (raro), rellenamos
		for len(p) < n {
			p += "x"
		}
		return p
	},

	// === Fecha personalizada con formato/rangos ===
	"random.Date": func(args map[string]string) string {
		layout := getArgOr(args, "format", "2006-01-02")
		startStr := getArgOr(args, "startDate", "1970-01-01")
		endStr := getArgOr(args, "endDate", time.Now().UTC().Format("2006-01-02"))

		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			start = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil || end.Before(start) {
			end = time.Now().UTC()
		}
		span := end.Unix() - start.Unix()
		if span <= 0 {
			return start.Format(layout)
		}
		sec := start.Unix() + rand.Int63n(span+1)
		return time.Unix(sec, 0).UTC().Format(layout)
	},
}
