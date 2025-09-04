package saga

import (
	"common/domain/customctx"
	"common/utils"
	"reflect"
)

type Payload map[string]any

type SAGA_Step interface {
	Call(ctx *customctx.CustomContext, payload utils.Result[Payload], allPayloads map[string]utils.Result[Payload]) utils.Result[Payload]
	Rollback(ctx *customctx.CustomContext) error
	Produce() string
}
type SAGA_Controller struct {
	Steps    []SAGA_Step
	Payloads map[string]utils.Result[Payload]
	PrevSaga *SAGA_Controller
}

func (c *SAGA_Controller) Executed(ctx *customctx.CustomContext) map[string]utils.Result[Payload] {
	allPayloads := make(map[string]utils.Result[Payload])
	var lastPayload utils.Result[Payload]
	for _, step := range c.Steps {

		result := step.Call(ctx, lastPayload, allPayloads)

		// Almacenar el resultado en allPayloads
		lastPayload = result

		name_step := step.Produce()
		if name_step == "" {
			t := reflect.TypeOf(step)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			name_step = t.Name()
		}

		allPayloads[name_step] = result

		if result.Err != nil {
			c.Rollback(ctx)
			if c.PrevSaga != nil {
				c.PrevSaga.Rollback(ctx)
			}
			break
		}
	}
	return allPayloads
}

func (c SAGA_Controller) Rollback(ctx *customctx.CustomContext) error {
	for i := len(c.Steps) - 1; i >= 0; i-- {
		err := c.Steps[i].Rollback(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c SAGA_Controller) Ok() bool {
	for _, p := range c.Payloads {
		if p.Err != nil {
			return false
		}
	}
	return true
}

func (c SAGA_Controller) Errors() []string {
	var errors []string
	for _, p := range c.Payloads {
		if p.Err != nil {
			errors = append(errors, p.Err.Error())
		}
	}
	return errors
}
