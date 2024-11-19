package alias

import (
	"context"
	"errors"
	"fmt"

	"github.com/otto8-ai/nah/pkg/name"
	"github.com/otto8-ai/nah/pkg/router"
	"github.com/otto8-ai/otto8/pkg/hash"
	v1 "github.com/otto8-ai/otto8/pkg/storage/apis/otto.otto8.ai/v1"
	"github.com/otto8-ai/otto8/pkg/system"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(ctx context.Context, c kclient.Client, obj v1.Aliasable, namespace string, name string) error {
	var errLookup error
	if namespace == "" {
		gvk, err := c.GroupVersionKindFor(obj.(kclient.Object))
		if err != nil {
			return err
		}
		errLookup = apierrors.NewNotFound(schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind,
		}, name)
	} else {
		errLookup = c.Get(ctx, router.Key(namespace, name), obj.(kclient.Object))
		if kclient.IgnoreNotFound(errLookup) != nil {
			return errLookup
		} else if errLookup == nil {
			return nil
		}
	}

	gvk, err := c.GroupVersionKindFor(obj.(kclient.Object))
	if err != nil {
		return err
	}

	var alias v1.Alias
	if err := c.Get(ctx, router.Key("", KeyFromScopeID(GetScope(gvk, obj), name)), &alias); apierrors.IsNotFound(err) {
		return errLookup
	} else if err != nil {
		return errors.Join(errLookup, err)
	} else if alias.Spec.TargetKind != gvk.Kind {
		return errLookup
	}

	return c.Get(ctx, router.Key(alias.Spec.TargetNamespace, alias.Spec.TargetName), obj.(kclient.Object))
}

func KeyFromScopeID(scope, id string) string {
	return system.AliasPrefix + hash.String(name.SafeHashConcatName(id, scope))[:8]
}

func GetScope(gvk schema.GroupVersionKind, obj v1.Aliasable) string {
	if scoped, ok := obj.(v1.AliasScoped); ok && scoped.GetAliasScope() != "" {
		return scoped.GetAliasScope()
	}

	return gvk.Kind
}

type GVKLookup interface {
	GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error)
}

type FromGVK schema.GroupVersionKind

func (f FromGVK) GroupVersionKindFor(_ runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind(f), nil
}

func Name(lookup GVKLookup, obj v1.Aliasable) (string, error) {
	id := obj.GetAliasName()
	if id == "" {
		return "", nil
	}
	runtimeObject, ok := obj.(runtime.Object)
	if !ok {
		return "", fmt.Errorf("object %T does not implement runtime.Object, can not lookup gvk", obj)
	}
	gvk, err := lookup.GroupVersionKindFor(runtimeObject)
	if err != nil {
		return "", err
	}
	return KeyFromScopeID(GetScope(gvk, obj), id), nil
}
