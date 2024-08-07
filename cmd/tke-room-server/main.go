package main

import (
	"context"
	"log/slog"
	"net/http"

	gamev1alpha1 "github.com/imroc/tke-room-manager/api/v1alpha1"
	"github.com/imroc/tke-room-manager/pkg/roomservice"
	"github.com/imroc/tke-room-manager/pkg/schemes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

func main() {
	config := ctrl.GetConfigOrDie()
	cls, err := cluster.New(config, func(o *cluster.Options) {
		o.Scheme = schemes.Scheme
	})
	if err != nil {
		panic(err)
	}
	gamev1alpha1.IndexField(cls.GetFieldIndexer())
	go func() {
		if err := cls.Start(context.Background()); err != nil {
			panic(err)
		}
	}()
	rs, err := roomservice.New(cls, schemes.Scheme)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	rs.AddHttpRoute(mux)
	slog.Info("starting server at :8000")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		panic(err)
	}
}
