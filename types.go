package main

import "k8s.io/apimachinery/pkg/runtime"

type Release struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
}

type Chart struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	AppVersion string `json:"appVersion"`
}

//
//type Vars struct {
//	Fullname string            `json:"fullname"`
//	Selector map[string]string `json:"selector"`
//	Labels   map[string]string `json:"labels"`
//}

type X struct {
	Release Release `json:"release"`
	Chart   Chart   `json:"chart"`
}

type ValuesModel struct {
	X       X                         `json:"x"`
	Objects map[string]runtime.Object `json:"objects"`
}

type ObjectModel struct {
	Key    string         `json:"key"`
	Object runtime.Object `json:"object"`
}
