package application

import (
	"fmt"
	"net/http"

	"github.com/AviralDixit-star/orders-api/handler"
	"github.com/AviralDixit-star/orders-api/repository/order"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (a *App) loadRouter() {
	fmt.Println("loadRouter")
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello"))
	})

	router.Route("/orders", a.loadOrderRoutes)
	a.router = router
}

func (a *App) loadOrderRoutes(router chi.Router) {
	fmt.Println("loadOrder")
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: a.rdb,
		},
	}
	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateByID)
	router.Delete("/{id}", orderHandler.DeleteByID)
}
