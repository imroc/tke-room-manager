package v1alpha1

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IndexField(indexer client.FieldIndexer) {
	indexer.IndexField(context.Background(), &Room{}, "status.ready", func(o client.Object) []string {
		ready := o.(*Room).Status.Ready
		return []string{fmt.Sprint(ready)}
	})
	indexer.IndexField(context.Background(), &Room{}, "status.idle", func(o client.Object) []string {
		idle := o.(*Room).Status.Idle
		return []string{fmt.Sprint(idle)}
	})
	indexer.IndexField(context.Background(), &Room{}, "spec.type", func(o client.Object) []string {
		tp := o.(*Room).Spec.Type
		return []string{tp}
	})
}
