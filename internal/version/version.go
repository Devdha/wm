package version

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func String() string {
	return Version + " (" + GitCommit + ") built " + BuildDate
}
