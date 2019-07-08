package ginflux

//RetentionPolicy ...
type RetentionPolicy struct {
	DBName        string
	RPName        string
	Duration      string
	Replication   int
	ShardDuration string
	Default       bool
}
