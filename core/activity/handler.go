package activity

import (
	"errors"
	"fmt"

	"github.com/infracloudio/msbotbuilder-go/schema"
)
type Handler interface {
	OnMessage(context *TurnContext) (schema.Activity, error)
	OnInvoke(context *TurnContext) (schema.Activity, error)
	OnConversationUpdate(context *TurnContext) (schema.Activity, error)
}
type HandlerFuncs struct {
	OnMessageFunc            func(turn *TurnContext) (schema.Activity, error)
	OnInvokeFunc             func(turn *TurnContext) (schema.Activity, error)
	OnConversationUpdateFunc func(turn *TurnContext) (schema.Activity, error)
}

func (r HandlerFuncs) OnMessage(turn *TurnContext) (schema.Activity, error) {
	if r.OnMessageFunc != nil {
		return r.OnMessageFunc(turn)
	}
	return schema.Activity{}, errors.New("No handler found for this activity type")
}

func (r HandlerFuncs) OnConversationUpdate(turn *TurnContext) (schema.Activity, error) {
	if r.OnConversationUpdateFunc != nil {
		return r.OnConversationUpdateFunc(turn)
	}
	return schema.Activity{}, errors.New("No handler found for this activity type")
}

func (r HandlerFuncs) OnInvoke(turn *TurnContext) (schema.Activity, error) {
	if r.OnInvokeFunc != nil {
		return r.OnInvokeFunc(turn)
	}
	return schema.Activity{}, errors.New("No handler found for this activity type")
}

func PrepareActivityContext(handler Handler, context *TurnContext) (schema.Activity, error) {
	switch context.Activity.Type {
	case schema.Message:
		return handler.OnMessage(context)
	case schema.Invoke:
		return handler.OnInvoke(context)
	case schema.ConversationUpdate:
		return handler.OnConversationUpdate(context)
	}
	return schema.Activity{}, fmt.Errorf("Activity type %s not supported yet", context.Activity.Type)
}
