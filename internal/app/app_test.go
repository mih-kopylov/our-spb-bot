package app

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestRunApplication(t *testing.T) {
	err := fx.ValidateApp(createApp("1", "2"))
	assert.NoError(t, err)
}

type ContainerLogConsumer struct {
	name   string
	logger *zap.Logger
}

func (c *ContainerLogConsumer) Accept(l testcontainers.Log) {
	c.logger.Info(string(l.Content),
		zap.String("container", c.name),
		zap.String("logType", l.LogType))
}

type TestFunc func(t *testing.T, mocks *Mocks)

func TestApp(t *testing.T) {
	mocks, err := NewMocks(t)
	if err != nil {
		t.Error(t)
	}

	mocks.BeforeAll(t)
	defer mocks.AfterAll(t)

	testFuncs := []TestFunc{ApplicationStarts}
	for _, testFunc := range testFuncs {
		runTestFunc(t, mocks, testFunc)
	}

}

func runTestFunc(t *testing.T, mocks *Mocks, testFunc TestFunc) {
	mocks.BeforeEach(t)
	defer mocks.AfterEach(t)

	funcName := getFuncName(testFunc)
	t.Run(funcName, func(t *testing.T) {
		testFunc(t, mocks)
	})
}

func getFuncName(testFunc TestFunc) string {
	funcForPC := runtime.FuncForPC(reflect.ValueOf(testFunc).Pointer())
	funcFullName := funcForPC.Name()
	return strings.TrimPrefix(filepath.Ext(funcFullName), ".")
}

func ApplicationStarts(t *testing.T, mocks *Mocks) {
	mocks.SetupGetMeMock(t)
	mocks.SetupSetMyCommandsMock(t)
	mocks.SetupGetUpdatesMock(t)
	app := fxtest.New(t, createApp("", ""))
	app.RequireStart()
	time.Sleep(3 * time.Second)
	app.RequireStop()
}
