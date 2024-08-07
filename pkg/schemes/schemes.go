package schemes

import (
	"k8s.io/apimachinery/pkg/runtime"

	gamev1alpha1 "github.com/imroc/tke-room-manager/api/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var Scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))

	utilruntime.Must(gamev1alpha1.AddToScheme(Scheme))
	// +kubebuilder:scaffold:scheme
}
