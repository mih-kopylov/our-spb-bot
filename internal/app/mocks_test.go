package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/mih-kopylov/our-spb-bot/internal/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/walkerus/go-wiremock"
	"go.uber.org/zap"
	"testing"
	"time"
)

type Mocks struct {
	WiremockContainer  testcontainers.Container
	WiremockPort       nat.Port
	WiremockClient     *wiremock.Client
	FirestoreContainer testcontainers.Container
	FirestorePort      nat.Port
}

func NewMocks(t *testing.T) (*Mocks, error) {
	ctx := context.Background()
	testcontainers.Logger = &TestContainersLogger{logger: log.NewLogger()}
	wiremockContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "wiremock/wiremock:latest-alpine",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.NewHostPortStrategy("8080/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Errorf("failed to start wiremock container: %v", err.Error())
	}

	wiremockContainer.FollowOutput(&ContainerLogConsumer{name: "wiremock", logger: log.NewLogger()})
	err = wiremockContainer.StartLogProducer(ctx)
	if err != nil {
		defer teardown(t, wiremockContainer)
	}

	wiremockPort, err := wiremockContainer.MappedPort(ctx, "8080/tcp")
	if err != nil {
		defer teardown(t, wiremockContainer)
		return nil, err
	}

	wiremockClient := wiremock.NewClient(fmt.Sprintf("http://localhost:%v", wiremockPort.Port()))

	firestoreContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mtlynch/firestore-emulator-docker:latest",
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				"FIRESTORE_PROJECT_ID": "ourspbbot",
			},
			WaitingFor: wait.ForLog("Dev App Server is now running"),
		},
		Started: true,
	})
	if err != nil {
		defer teardown(t, wiremockContainer)
		return nil, err
	}

	firestoreContainer.FollowOutput(&ContainerLogConsumer{name: "firestore", logger: log.NewLogger()})
	err = firestoreContainer.StartLogProducer(ctx)
	if err != nil {
		defer teardown(t, wiremockContainer, firestoreContainer)
	}

	firestorePort, err := firestoreContainer.MappedPort(ctx, "8080/tcp")
	if err != nil {
		defer teardown(t, wiremockContainer, firestoreContainer)
		return nil, err
	}

	return &Mocks{
		WiremockContainer:  wiremockContainer,
		WiremockPort:       wiremockPort,
		WiremockClient:     wiremockClient,
		FirestoreContainer: firestoreContainer,
		FirestorePort:      firestorePort,
	}, nil
}

func (m *Mocks) BeforeAll(t *testing.T) {
	t.Setenv("TELEGRAM_API_TOKEN", "TELEGRAM_API_TOKEN")
	t.Setenv("TELEGRAM_API_ENDPOINT", fmt.Sprintf("http://localhost:%v/bot%%s/%%s", m.WiremockPort.Port()))
	t.Setenv("OURSPB_API_ENDPOINT", fmt.Sprintf("http://localhost:%v", m.WiremockPort.Port()))
	t.Setenv("OURSPB_CLIENT_ID", "OURSPB_CLIENT_ID")
	t.Setenv("OURSPB_SECRET", "OURSPB_SECRET")
	t.Setenv("FIRESTORE_EMULATOR_HOST", fmt.Sprintf("localhost:%v", m.FirestorePort.Port()))
	t.Setenv("FIREBASE_SERVICE_ACCOUNT", base64.StdEncoding.EncodeToString([]byte("FIREBASE_SERVICE_ACCOUNT")))
	t.Setenv("SENDER_SLEEP_DURATION", "1s")
}

func teardown(t *testing.T, containers ...testcontainers.Container) {
	t.Log("terminating containers")
	for _, container := range containers {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("failed to terminate container: %s", err.Error())
		}
	}
}

func (m *Mocks) AfterAll(t *testing.T) {
	teardown(t, m.WiremockContainer, m.FirestoreContainer)
}

func (m *Mocks) BeforeEach(_ *testing.T) {

}

func (m *Mocks) AfterEach(t *testing.T) {
	err := m.WiremockClient.Clear()
	if err != nil {
		t.Error(err)
	}
}

func (m *Mocks) SetupGetMeMock(t *testing.T) {
	err := m.WiremockClient.StubFor(wiremock.Post(wiremock.URLPathMatching("/bot.+/getMe")).
		WillReturnJSON(
			map[string]any{
				"ok": true,
				"result": map[string]any{
					"id":     1,
					"is_bot": true,
				},
			},
			map[string]string{"Content-Type": "application/json"},
			200,
		))
	if err != nil {
		t.Error(err)
	}
}

func (m *Mocks) SetupSetMyCommandsMock(t *testing.T) {
	err := m.WiremockClient.StubFor(wiremock.Post(wiremock.URLPathMatching("/bot.+/setMyCommands")).
		WillReturnJSON(
			map[string]any{
				"ok": true,
			},
			map[string]string{"Content-Type": "application/json"},
			200,
		))
	if err != nil {
		t.Error(err)
	}
}
func (m *Mocks) SetupGetUpdatesMock(t *testing.T) {
	err := m.WiremockClient.StubFor(wiremock.Post(wiremock.URLPathMatching("/bot.+/getUpdates")).
		WithFixedDelayMilliseconds(time.Second).
		WillReturnJSON(
			map[string]any{
				"ok":     true,
				"result": []any{},
			},
			map[string]string{"Content-Type": "application/json"},
			200,
		))
	if err != nil {
		t.Error(err)
	}
}

type TestContainersLogger struct {
	logger *zap.Logger
}

func (t *TestContainersLogger) Printf(format string, v ...interface{}) {
	t.logger.Sugar().Infof(format, v...)
}
