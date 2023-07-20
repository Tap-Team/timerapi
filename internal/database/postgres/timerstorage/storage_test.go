package timerstorage_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Tap-Team/timerapi/internal/database/postgres/timerstorage"
	"github.com/Tap-Team/timerapi/pkg/postgres"
)

var (
	testTimerStorage *timerstorage.Storage
)

func TestMain(m *testing.M) {
	os.Setenv("TZ", "UTC")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	postgres, terminate, err := postgres.NewContainer(ctx, postgres.DEFAULT_MIGRATION_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer terminate(ctx)
	testTimerStorage = timerstorage.New(postgres)
	m.Run()
}
