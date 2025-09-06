package placeholder

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ==== Contexto ====
type MockContext struct {
	PathParams map[string]string
	Query      map[string]string
	Headers    map[string]string
	Body       map[string]any
}

// Utilidad: obtener arg (si no existe, default)
func getArgOr(args map[string]string, key, def string) string {
	if v, ok := args[key]; ok && v != "" {
		return v
	}
	return def
}

// ==== Parsing de llamadas: "random.Date(format:'2006-01-02', startDate:'1980-01-01')" ====
var funcCallRe = regexp.MustCompile(`^\s*([A-Za-z0-9\.\_\-]+)\s*(?:\((.*)\))?\s*$`)

// parseFuncCall retorna nombre y mapa de args. Si no hay paréntesis, args=nil
func parseFuncCall(expr string) (name string, args map[string]string) {
	m := funcCallRe.FindStringSubmatch(expr)
	if m == nil {
		// no matchea, devolver tal cual
		return expr, nil
	}
	name = m[1]
	raw := strings.TrimSpace(m[2])
	if raw == "" {
		return name, nil
	}
	args = parseArgs(raw)
	return name, args
}

// parseArgs soporta formato: key:'val', key:"val", key: val (sin comillas, sin comas internas)
func parseArgs(s string) map[string]string {
	out := map[string]string{}
	// split por coma de primer nivel (simple)
	parts := splitTopLevelComma(s)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		// quitar comillas simples/dobles si existen
		v = trimQuotes(v)
		out[k] = v
	}
	return out
}

func splitTopLevelComma(s string) []string {
	// aquí basta con dividir por comas; si quisieras anidación, tocaría parser más robusto
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, p)
	}
	return out
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// Reemplaza placeholders {{...}} dentro de strings
func replacePlaceholders(input string, ctx MockContext) string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"))

		// ---- Random (map extensible + args opcionales) ----
		name, args := parseFuncCall(key)
		if gen, ok := randomGenerators[name]; ok {
			// args puede ser nil; los generadores esperan map[string]string (nil ok)
			return gen(args)
		}

		// ---- Path params ----
		if strings.HasPrefix(name, "path.") {
			field := strings.TrimPrefix(name, "path.")
			return ctx.PathParams[field]
		}

		// ---- Query params ----
		if strings.HasPrefix(name, "query.") {
			field := strings.TrimPrefix(name, "query.")
			return ctx.Query[field]
		}

		// ---- Headers ----
		if strings.HasPrefix(name, "headers.") {
			field := strings.TrimPrefix(name, "headers.")
			return ctx.Headers[field]
		}

		// ---- Body (soporta subcampos body.a.b.c) ----
		if strings.HasPrefix(name, "body.") {
			field := strings.TrimPrefix(name, "body.")
			parts := strings.Split(field, ".")
			val := getNested(ctx.Body, parts)
			switch v := val.(type) {
			case string:
				// Si el valor del body trae otro placeholder, resolver en cascada
				return replacePlaceholders(v, ctx)
			default:
				if v == nil {
					return ""
				}
				b, _ := json.Marshal(v)
				return string(b)
			}
		}

		// si no coincide nada, regresamos el placeholder intacto
		return match
	})
}

// Busca un valor anidado en map[string]any
func getNested(m map[string]any, keys []string) any {
	if len(keys) == 0 {
		return nil
	}
	val, ok := m[keys[0]]
	if !ok {
		return nil
	}
	if len(keys) == 1 {
		return val
	}
	if sub, ok := val.(map[string]any); ok {
		return getNested(sub, keys[1:])
	}
	return nil
}

// Resuelve recursivamente maps y arrays
func resolvePlaceholdersDeep(node any, ctx MockContext) any {
	switch v := node.(type) {
	case string:
		return replacePlaceholders(v, ctx)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = resolvePlaceholdersDeep(val, ctx)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = resolvePlaceholdersDeep(val, ctx)
		}
		return out
	default:
		return v
	}
}

type PlaceholderController struct {
}

func NewPlaceholderController() *PlaceholderController {
	return &PlaceholderController{}
}

func (c *PlaceholderController) Resolve(ctx MockContext, input map[string]any) (any, error) {

	resolved := resolvePlaceholdersDeep(input, ctx)
	return resolved, nil
}
