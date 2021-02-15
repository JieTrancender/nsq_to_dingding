package instance

// Settings contains basic settings for any consumer to pass into GenRootCmd
type Settings struct {
	Name        string
	IndexPrefix string
	Version     string
	// RunFlags cobra.FlagSet

	Umask *int
}
