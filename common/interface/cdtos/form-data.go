package cdtos

import (
	"common/domain/customctx"
	"common/utils"
	"common/utils/cerrs"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var ErrNotMultipart = errors.New("content-type debe ser multipart/form-data")

// BindFormData bindea un multipart/form-data al tipo T.
// Tags:
//
//	form:"key"        -> valores (string/bool/int/float/time) y sus slices
//	file:"key"        -> *multipart.FileHeader o []*multipart.FileHeader
//	required:"true"   -> obligatorio
//	default:"value"   -> valor por defecto cuando falta
//	time_layout:"..." -> layout para time.Parse (default RFC3339)
func BindFormData[T any](c *gin.Context, cc *customctx.CustomContext) utils.Response[T] {
	if !strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		return utils.Response[T]{
			StatusCode: http.StatusBadRequest,
			Error: cc.NewError(
				cerrs.NewCustomError(
					http.StatusBadRequest,
					ErrNotMultipart.Error(),
					"bind_form_data.not_multipart",
				),
			),
		}
	}
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return utils.Response[T]{
			StatusCode: http.StatusBadRequest,
			Error: cc.NewError(
				cerrs.NewCustomError(
					http.StatusBadRequest,
					fmt.Errorf("parse multipart: %w", err).Error(),
					"bind_form_data.multipart_form_empty",
				),
			),
		}
	}
	mf := c.Request.MultipartForm
	if mf == nil {
		return utils.Response[T]{
			StatusCode: http.StatusBadRequest,
			Error: cc.NewError(
				cerrs.NewCustomError(
					http.StatusBadRequest,
					"multipart form vacío",
					"bind_form_data.multipart_form_empty",
				),
			),
		}
	}

	var out T
	rv := reflect.ValueOf(&out).Elem()
	rt := rv.Type()

	var errs []string

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		formKey := tagOr(sf, "form", strings.ToLower(sf.Name))
		// AHORA usamos binding:"required"
		req := tagHas(sf, "binding", "required")
		def := sf.Tag.Get("default")
		timeFmt := tagOr(sf, "time_format", time.RFC3339)

		ft := sf.Type

		switch {
		// ---- Archivos: usan form:"..." y se distinguen por tipo ----
		case isFileHeaderPtr(ft):
			files := mf.File[formKey]
			if len(files) > 0 {
				fv.Set(reflect.ValueOf(files[0]))
			} else if req {
				errs = append(errs, fmt.Sprintf("archivo requerido: %q", formKey))
			}

		case isFileHeaderSlice(ft):
			files := mf.File[formKey]
			if len(files) > 0 {
				fv.Set(reflect.ValueOf(files))
			} else if req {
				errs = append(errs, fmt.Sprintf("archivos requeridos: %q", formKey))
			}

		// ---- Valores ----
		default:
			values := mf.Value[formKey]
			if len(values) == 0 && def != "" {
				values = []string{def}
			}
			if req && len(values) == 0 {
				errs = append(errs, fmt.Sprintf("campo requerido: %q", formKey))
				continue
			}
			if len(values) == 0 {
				continue
			}

			if ft.Kind() == reflect.Slice {
				elem := ft.Elem()
				slice, err := parseSlice(values, elem, timeFmt)
				if err != nil {
					errs = append(errs, fmt.Sprintf("%s: %v", formKey, err))
					continue
				}
				fv.Set(slice)
				continue
			}

			val, err := parseScalar(values[0], ft, timeFmt)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", formKey, err))
				continue
			}
			fv.Set(val)
		}
	}

	if len(errs) > 0 {
		return utils.Response[T]{
			StatusCode: http.StatusBadRequest,
			Error: cc.NewError(
				cerrs.NewCustomError(
					http.StatusBadRequest,
					strings.Join(errs, "; "),
					"bind_form_data.validation_errors",
				),
			),
		}
	}
	return utils.Response[T]{Data: out}
}

// ---------- helpers ----------

func tagOr(sf reflect.StructField, key, fallback string) string {
	if v := sf.Tag.Get(key); v != "" {
		return v
	}
	return fallback
}

func tagBool(sf reflect.StructField, key string) bool {
	return strings.EqualFold(sf.Tag.Get(key), "true")
}

func chooseKey(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func isFileHeaderPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem() == reflect.TypeOf(multipart.FileHeader{})
}
func isFileHeaderSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice &&
		t.Elem().Kind() == reflect.Ptr &&
		t.Elem().Elem() == reflect.TypeOf(multipart.FileHeader{})
}

func isTime(t reflect.Type) bool {
	return t.PkgPath() == "time" && t.Name() == "Time"
}

func parseSlice(values []string, elem reflect.Type, layout string) (reflect.Value, error) {
	s := reflect.MakeSlice(reflect.SliceOf(elem), 0, len(values))
	for _, v := range values {
		x, err := parseScalar(v, elem, layout)
		if err != nil {
			return reflect.Value{}, err
		}
		s = reflect.Append(s, x)
	}
	return s, nil
}

func parseScalar(s string, t reflect.Type, layout string) (reflect.Value, error) {
	switch {
	case isTime(t):
		ts, err := time.Parse(layout, s)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("time inválido (%s): %w", layout, err)
		}
		return reflect.ValueOf(ts), nil
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(s), nil
	case reflect.Bool:
		b, err := strconv.ParseBool(strings.TrimSpace(s))
		if err != nil {
			return reflect.Value{}, fmt.Errorf("bool inválido: %q", s)
		}
		return reflect.ValueOf(b), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(strings.TrimSpace(s), 10, t.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("int inválido: %q", s)
		}
		v := reflect.New(t).Elem()
		v.SetInt(i)
		return v, nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(strings.TrimSpace(s), t.Bits())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("float inválido: %q", s)
		}
		v := reflect.New(t).Elem()
		v.SetFloat(f)
		return v, nil
	default:
		return reflect.Value{}, fmt.Errorf("tipo no soportado: %s", t)
	}
}

func tagHas(sf reflect.StructField, key, want string) bool {
	// p.ej. binding:"required,min=1" -> detecta "required"
	raw := sf.Tag.Get(key)
	if raw == "" {
		return false
	}
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), want) {
			return true
		}
	}
	return false
}
