package xhealth

import (
	"encoding/json"
	"fmt"

	health "github.com/InVisionApp/go-health"
	"github.com/xmidt-org/themis/service"
)

// Checker is an implementation of service.Checker.  Instances of this type
// can be wired up to service discovery for things like TTL checks.
type Checker struct {
	h      health.IHealth
	custom map[string]interface{}
}

func (c *Checker) generateMessage(states map[string]health.State, failed bool) map[string]interface{} {
	message := map[string]interface{}{
		"details": states,
	}

	if failed {
		message["status"] = "failed"
	} else {
		message["status"] = "ok"
	}

	for k, v := range c.custom {
		if k != "status" && k != "details" {
			message[k] = v
		}
	}

	return message
}

// Check marshals the InvisionApp state in a similar manner to the JSON handler and makes that available as the status message.
// The only difference here is that the Message is always JSON, even in cases of marshalling errors.
// See https://github.com/InVisionApp/go-health/blob/master/handlers/handlers.go#L45
func (c *Checker) Check() service.CheckStatus {
	states, failed, err := c.h.State()
	if err != nil {
		return service.CheckStatus{
			State:   service.Fail,
			Message: fmt.Sprintf(`{"message": "Unable to fetch states", "error": "%v"}`, err),
		}
	}

	message := c.generateMessage(states, failed)
	data, err := json.Marshal(message)
	if err != nil {
		return service.CheckStatus{
			State:   service.Fail,
			Message: fmt.Sprintf(`{"message": "Failed to marshal state data", "error": "%v"}`, err),
		}
	}

	state := service.Pass
	if failed {
		state = service.Fail
	}

	return service.CheckStatus{
		State:   state,
		Message: string(data),
	}
}
