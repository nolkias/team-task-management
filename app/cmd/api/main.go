package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"teamtask/internal/breaker"
	"teamtask/internal/cache"
	"teamtask/internal/config"
	"teamtask/internal/jwtutil"
	"teamtask/internal/metrics"
	"teamtask/internal/repository/mysql"
	"teamtask/internal/service"
	transporthttp "teamtask/internal/transport/http"
)

func main() {
	yamlPath, envPath := config.DefaultPaths()
	cfg, err := config.Load(yamlPath, envPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := mysql.NewPool(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := waitForDB(db, 30*time.Second); err != nil {
		log.Fatalf("database not reachable: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	redisClient := cache.NewClient(cfg.Redis)
	defer redisClient.Close()

	metrics.Register()

	userRepo := mysql.NewUserRepo(db)
	teamRepo := mysql.NewTeamRepo(db)
	taskRepo := mysql.NewTaskRepo(db)
	historyRepo := mysql.NewTaskHistoryRepo(db)

	issuer := jwtutil.NewIssuer(cfg.JWT.Secret, cfg.JWT.Expiry())
	taskListCache := cache.NewTaskListCache(redisClient, cfg.Cache.TaskListTTL())
	rateLimiter := cache.NewRateLimiter(redisClient, cfg.RateLimit.RequestsPerMinute, time.Minute)

	emailService := breaker.NewEmailServiceBreaker(service.NewMockEmailService())

	authService := service.NewAuthService(userRepo, issuer)
	teamService := service.NewTeamService(teamRepo, userRepo, emailService)
	taskService := service.NewTaskService(taskRepo, teamRepo, taskListCache)

	handlers := transporthttp.Handlers{
		Auth:  transporthttp.NewAuthHandler(authService),
		Team:  transporthttp.NewTeamHandler(teamService),
		Task:  transporthttp.NewTaskHandler(taskService, historyRepo, cfg.Pagination.DefaultPageSize, cfg.Pagination.MaxPageSize),
		Admin: transporthttp.NewAdminHandler(taskRepo, teamRepo),
	}

	router := transporthttp.NewRouter(handlers, issuer, rateLimiter)
	server := transporthttp.NewServer(cfg.Server.Port, router)

	go func() {
		log.Printf("server listening on port %d", cfg.Server.Port)
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received, draining in-flight requests")
	if err := server.Shutdown(cfg.Server.ShutdownTimeout()); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("server stopped cleanly")
}

func waitForDB(db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var err error
	for time.Now().Before(deadline) {
		if err = db.Ping(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return err
}

func runMigrations(db *sql.DB) error {
	driver, err := mysqlmigrate.WithInstance(db, &mysqlmigrate.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
