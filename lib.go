package main

import (
	"github.com/gobuffalo/flect"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

func NG(apiVersion, kind, name, fullName string, histogram map[schema.GroupKind]int) (string, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return "", err
	}
	gk := schema.GroupKind{
		Group: gv.Group,
		Kind:  kind,
	}

	groupPrefix := gv.Group
	groupPrefix = strings.TrimSuffix(groupPrefix, ".k8s.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".kubernetes.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	if count, ok := histogram[gk]; ok && count > 1 {
		nameSuffix = strings.TrimPrefix(name, fullName)
		nameSuffix = strings.Replace(nameSuffix, ".", "-", -1)
		nameSuffix = strings.Trim(nameSuffix, "-")
		nameSuffix = flect.Pascalize(nameSuffix)
	}

	return flect.Camelize(groupPrefix + kind + nameSuffix), nil
}
