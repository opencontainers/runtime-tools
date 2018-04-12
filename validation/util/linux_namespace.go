package util

// ProcNamespaces defines a list of namespaces to be found under /proc/*/ns/.
// NOTE: it is not the same as generate.Namespaces, because of naming
// mismatches like "mnt" vs "mount" or "net" vs "network".
var ProcNamespaces = []string{
	"cgroup",
	"ipc",
	"mnt",
	"net",
	"pid",
	"user",
	"uts",
}
