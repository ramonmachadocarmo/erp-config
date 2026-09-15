package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"erp/pkg/config"
	"erp/pkg/httpserver"
	"erp/pkg/postgres"
	geoadapter "erp/services/config-service/internal/adapters/geo"
	httpadapter "erp/services/config-service/internal/adapters/http"
	pgadapter "erp/services/config-service/internal/adapters/postgres"
	"erp/services/config-service/internal/application"
	"erp/services/config-service/migrations"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := postgres.EnsureDatabase(ctx, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DB); err != nil {
		log.Fatal(err)
	}
	pool, err := postgres.Connect(ctx, cfg.Postgres.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool, migrations.FS, "."); err != nil {
		log.Fatal(err)
	}

	svc := application.New(
		pgadapter.NewUnitRepo(pool),
		pgadapter.NewPersonRepo(pool),
		pgadapter.NewCustomerRepo(pool),
		pgadapter.NewSupplierRepo(pool),
		pgadapter.NewSettingRepo(pool),
		pgadapter.NewMethodRepo(pool),
		pgadapter.NewTermRepo(pool),
		pgadapter.NewCenterRepo(pool),
		pgadapter.NewVehicleRepo(pool),
		pgadapter.NewCompanyRepo(pool),
		geoadapter.New(),
		postgres.NewSequence(pool, "unit"),
		postgres.NewSequence(pool, "payment_method"),
		postgres.NewSequence(pool, "payment_term"),
		postgres.NewSequence(pool, "distribution_center"),
		postgres.NewSequence(pool, "delivery_vehicle"),
	)
	engine := httpserver.New(cfg.ServiceName)
	httpadapter.New(svc).Register(engine, httpserver.JWT(cfg.JWTSecret, cfg.JWTIssuer))

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: engine}
	go func() {
		log.Printf("%s listening on %s", cfg.ServiceName, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}
