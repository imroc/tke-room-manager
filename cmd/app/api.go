package app

import (
	"net/http"

	"github.com/imroc/tke-room-manager/internal/sidecar/roomservice"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func runApiServer(mgr manager.Manager) error {
	addr := viper.GetString(apiBindAddress)
	rs, err := roomservice.New(mgr, mgr.GetScheme())
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	rs.AddHttpRoute(mux)
	return http.ListenAndServe(addr, mux)
}
