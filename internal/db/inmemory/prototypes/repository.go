package inmemory

import (
	"common/domain/criteria"
	"common/domain/customctx"
	"common/utils"
	"common/utils/cerrs"
	"context"
	"encoding/json"
	"mocky/internal/db/mongo/prototypes"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	// ajusta este import al paquete real de tus modelos
)

// entrada con TTL
type entry struct {
	model     prototypes.PrototypeModel
	expiresAt time.Time
}

type InMemoryPrototypesRepository struct {
	mu        sync.RWMutex
	store     map[string]entry
	byPathKey map[string]string
	toList    func(prototypes.PrototypeModel) prototypes.PrototypeListModel
	ttl       time.Duration // p.ej., 5 * time.Minute
}

// NewInMemoryPrototypesRepository crea un repo con TTL fijo por entrada.
// ttl: tiempo de vida de cada registro (si <=0 usa 5 min).
func NewInMemoryPrototypesRepository(
	toList func(prototypes.PrototypeModel) prototypes.PrototypeListModel,
	ttl time.Duration,
) *InMemoryPrototypesRepository {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &InMemoryPrototypesRepository{
		store:     make(map[string]entry),
		byPathKey: make(map[string]string),
		toList:    toList,
		ttl:       ttl,
	}
}

// ===================== Helpers =====================

func keyFor(method, urlPath string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + "\n" + strings.TrimSpace(urlPath)
}

func setByDottedPath(root map[string]any, path string, val any) {
	parts := strings.Split(path, ".")
	m := root
	for i, p := range parts {
		if i == len(parts)-1 {
			m[p] = val
			return
		}
		next, ok := m[p]
		if !ok {
			nm := map[string]any{}
			m[p] = nm
			m = nm
			continue
		}
		switch typed := next.(type) {
		case map[string]any:
			m = typed
		default:
			nm := map[string]any{}
			m[p] = nm
			m = nm
		}
	}
}

func (r *InMemoryPrototypesRepository) put(id string, m prototypes.PrototypeModel) {
	r.store[id] = entry{
		model:     m,
		expiresAt: time.Now().Add(r.ttl),
	}
	r.byPathKey[keyFor(m.Request.Method, m.Request.UrlPath)] = id
}

func (r *InMemoryPrototypesRepository) getIfAliveByID(id string) (prototypes.PrototypeModel, bool) {
	e, ok := r.store[id]
	if !ok {
		return prototypes.PrototypeModel{}, false
	}
	if time.Now().After(e.expiresAt) {
		// caducado: limpiar
		key := keyFor(e.model.Request.Method, e.model.Request.UrlPath)
		delete(r.byPathKey, key)
		delete(r.store, id)
		return prototypes.PrototypeModel{}, false
	}
	return e.model, true
}

func (r *InMemoryPrototypesRepository) getIfAliveByKey(method, urlPath string) (prototypes.PrototypeModel, bool) {
	id, ok := r.byPathKey[keyFor(method, urlPath)]
	if !ok {
		return prototypes.PrototypeModel{}, false
	}
	return r.getIfAliveByID(id)
}

// ================= Implementación RepositoryPrototypes =================

func (r *InMemoryPrototypesRepository) Save(ctx context.Context, document prototypes.PrototypeModel) utils.Result[string] {
	id := primitive.NewObjectID().Hex()
	document.ID = id
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	}
	document.UpdatedAt = time.Now()

	r.mu.Lock()
	r.put(id, document)
	r.mu.Unlock()

	return utils.Result[string]{Data: id}
}

func (r *InMemoryPrototypesRepository) SaveWithID(ctx context.Context, id string, document prototypes.PrototypeModel) utils.Result[string] {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusBadRequest, "id inválido (no es ObjectID hex)", "inmemory.save_with_id")}
	}
	document.ID = id
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	}
	document.UpdatedAt = time.Now()

	r.mu.Lock()
	r.put(id, document)
	r.mu.Unlock()

	return utils.Result[string]{Data: id}
}

func (r *InMemoryPrototypesRepository) Update(ctx context.Context, entity prototypes.PrototypeModel) error {
	if entity.ID == "" {
		return cerrs.NewCustomError(http.StatusBadRequest, "entity.ID vacío", "inmemory.update")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.getIfAliveByID(entity.ID)
	if !ok {
		return cerrs.NewCustomError(http.StatusNotFound, "no existe el prototipo", "inmemory.update")
	}

	entity.UpdatedAt = time.Now()
	r.put(entity.ID, entity)
	return nil
}

func (r *InMemoryPrototypesRepository) UpdateFields(ctx context.Context, id string, updates map[string]interface{}) utils.Result[prototypes.PrototypeModel] {
	r.mu.Lock()
	defer r.mu.Unlock()

	cur, ok := r.getIfAliveByID(id)
	if !ok {
		return utils.Result[prototypes.PrototypeModel]{Err: cerrs.NewCustomError(http.StatusNotFound, "no se encontró el prototipo", "inmemory.update_fields")}
	}

	// model -> map
	var m map[string]any
	b, _ := json.Marshal(cur)
	_ = json.Unmarshal(b, &m)

	for k, v := range updates {
		if strings.Contains(k, ".") {
			setByDottedPath(m, k, v)
		} else {
			m[k] = v
		}
	}

	// map -> model
	var out prototypes.PrototypeModel
	nb, _ := json.Marshal(m)
	if err := json.Unmarshal(nb, &out); err != nil {
		return utils.Result[prototypes.PrototypeModel]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "inmemory.update_fields.unmarshal")}
	}

	out.ID = id
	out.UpdatedAt = time.Now()
	if out.CreatedAt.IsZero() {
		out.CreatedAt = cur.CreatedAt
	}
	r.put(id, out)

	return utils.Result[prototypes.PrototypeModel]{Data: out}
}

func (r *InMemoryPrototypesRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.store[id]
	if !ok {
		return cerrs.NewCustomError(http.StatusNotFound, "no se encontró el prototipo", "inmemory.delete")
	}
	key := keyFor(e.model.Request.Method, e.model.Request.UrlPath)
	delete(r.byPathKey, key)
	delete(r.store, id)
	return nil
}

func (r *InMemoryPrototypesRepository) Find(ctx context.Context, id string) utils.Result[prototypes.PrototypeModel] {
	r.mu.Lock() // Lock para poder purgar si expiró
	defer r.mu.Unlock()

	m, ok := r.getIfAliveByID(id)
	if !ok {
		return utils.Result[prototypes.PrototypeModel]{Err: cerrs.NewCustomError(http.StatusNotFound, "no se encontró el prototipo", "inmemory.find")}
	}
	return utils.Result[prototypes.PrototypeModel]{Data: m}
}

func (r *InMemoryPrototypesRepository) FindAll(ctx context.Context) utils.Result[[]prototypes.PrototypeListModel] {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	list := make([]prototypes.PrototypeListModel, 0, len(r.store))
	for id, e := range r.store {
		if now.After(e.expiresAt) {
			key := keyFor(e.model.Request.Method, e.model.Request.UrlPath)
			delete(r.byPathKey, key)
			delete(r.store, id)
			continue
		}
		if r.toList != nil {
			list = append(list, r.toList(e.model))
		}
	}
	return utils.Result[[]prototypes.PrototypeListModel]{Data: list}
}

func (r *InMemoryPrototypesRepository) Matching(cr criteria.Criteria, _ string, offset int, limit int) utils.Result[[]prototypes.PrototypeListModel] {
	var wantURL, wantMethod *string
	if cr.Filters.Get != nil {
		for _, f := range cr.Filters.Get() {
			switch strings.ToLower(string(f.Field)) {
			case "request.urlpath":
				if s, ok := f.Value.(string); ok {
					wantURL = &s
				}
			case "request.method":
				if s, ok := f.Value.(string); ok {
					wantMethod = &s
				}
			}
		}
	}

	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]prototypes.PrototypeListModel, 0, len(r.store))
	for id, e := range r.store {
		if now.After(e.expiresAt) {
			key := keyFor(e.model.Request.Method, e.model.Request.UrlPath)
			delete(r.byPathKey, key)
			delete(r.store, id)
			continue
		}
		if wantURL != nil && e.model.Request.UrlPath != *wantURL {
			continue
		}
		if wantMethod != nil && !strings.EqualFold(e.model.Request.Method, *wantMethod) {
			continue
		}
		if r.toList != nil {
			items = append(items, r.toList(e.model))
		}
	}

	// paginación
	start := offset
	if start < 0 {
		start = 0
	}
	if start > len(items) {
		start = len(items)
	}
	end := len(items)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	return utils.Result[[]prototypes.PrototypeListModel]{Data: items[start:end]}
}

func (r *InMemoryPrototypesRepository) GetByPath(cc *customctx.CustomContext, urlPath string, method string) utils.Result[prototypes.PrototypeModel] {
	select {
	case <-cc.Context().Done():
		return utils.Result[prototypes.PrototypeModel]{Err: cerrs.NewCustomError(http.StatusRequestTimeout, "context canceled", "inmemory.get_by_path")}
	default:
	}

	r.mu.Lock() // Lock para poder purgar si caducó
	defer r.mu.Unlock()

	m, ok := r.getIfAliveByKey(method, urlPath)
	if !ok {
		return utils.Result[prototypes.PrototypeModel]{Err: cerrs.NewCustomError(http.StatusNotFound, "dont have prototype for this path: "+urlPath+" and method: "+method, "inmemory.get_by_path")}
	}
	return utils.Result[prototypes.PrototypeModel]{Data: m}
}

func (r *InMemoryPrototypesRepository) SaveOrUpdate(cc *customctx.CustomContext, document prototypes.PrototypeModel) utils.Result[string] {
	if document.Request.BodySchema != nil && document.Request.BodySchema.TypeSchema == "" {
		document.Request.BodySchema = nil
	}

	existing := r.GetByPath(cc, document.Request.UrlPath, document.Request.Method)
	if existing.Err != nil {
		// nuevo
		document.CreatedAt = time.Now()
		document.UpdatedAt = time.Now()
		return r.Save(cc.Context(), document)
	}

	// reemplazo preservando ID y CreatedAt
	document.ID = existing.Data.ID
	if document.CreatedAt.IsZero() {
		document.CreatedAt = existing.Data.CreatedAt
	}
	document.UpdatedAt = time.Now()

	r.mu.Lock()
	r.put(document.ID, document)
	r.mu.Unlock()

	return utils.Result[string]{Data: document.ID}
}
