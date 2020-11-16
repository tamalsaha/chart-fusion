# chart-fusion
Fuse Kubernetes YAMLs into Zero Template Helm Charts

```
foldername:
  a.ymal
  b.yaml



Get names of each resource
  remove prefix
  generate key
    - "" => camel({apiGroup}){Kind}, sanke(camel({apiGroup}){Kind}).yaml
    - !"" => {name}, sanke({name}).yaml
  values:
    key:
      apiVersion:
      kind: 
      metdata:
      spec:
      # remove status
```

```
helm template kubedb kubedb > /home/tamal/go/src/github.com/tamalsaha/chart-fusion/kubedb/chart.yaml
```

render-gotpl --template=templates/object.yaml --data=sample.json > report/README.md


