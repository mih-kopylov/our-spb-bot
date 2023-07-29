package state

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormStringValue(t *testing.T) {
	state := BaseUserState{}
	state.SetFormField("key", "value")
	actual := state.GetStringFormField("key")
	assert.Equal(t, "value", actual)
}

func TestFormIntValue(t *testing.T) {
	state := BaseUserState{}
	state.SetFormField("key", 15)
	actual := state.GetStringFormField("key")
	assert.Equal(t, "", actual)
}

func TestFormBoolValue(t *testing.T) {
	state := BaseUserState{}
	state.SetFormField("key", true)
	actual := state.GetStringFormField("key")
	assert.Equal(t, "", actual)
}

func TestFormNoForm(t *testing.T) {
	state := BaseUserState{}
	actual := state.GetStringFormField("key")
	assert.Equal(t, "", actual)
}

func TestFormNoValue(t *testing.T) {
	state := BaseUserState{}
	state.SetFormField("key2", "value")
	actual := state.GetStringFormField("key")
	assert.Equal(t, "", actual)
}

func TestFormStringSliceValue(t *testing.T) {
	state := BaseUserState{}
	state.AddValueToStringSlice("key", "value")
	actual := state.GetStringSlice("key")
	assert.Equal(t, []string{"value"}, actual)
}

func TestFormStringSliceSameValue(t *testing.T) {
	state := BaseUserState{}
	state.AddValueToStringSlice("key", "value")
	state.AddValueToStringSlice("key", "value")
	actual := state.GetStringSlice("key")
	assert.Equal(t, []string{"value", "value"}, actual)
}
