package slice

func Dedup(list []string) []string {
	allVals := make(map[string]bool)
	d := []string{}
	for _, s := range list {
		if _, ok := allVals[s]; !ok {
			allVals[s] = true
			d = append(d, s)
		}
	}
	return d
}
