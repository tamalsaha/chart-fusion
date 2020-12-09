
func detectmode(model) {
	if model.mysql.spec.shard != nil {
		return "Shard"
	} else model.spec.replicas > 1 {
		return "replicated"
	}
	return "standalone"
}

