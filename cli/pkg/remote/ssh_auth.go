package remote

func DefaultPrivateKeys() []string {
	return []string{
		"~/.ssh/id_rsa",
		"~/.ssh/id_dsa",
		"~/.ssh/google_compute_engine",
	}
}
