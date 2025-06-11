package typesense_healthcheck

type Config struct {
	LogLevel        int    `env:"LOG_LEVEL" envDefault:"0"`
	Logo            string `env:"LOGO" envDefault:"https://akyriako.github.io/typesense-operator-docs/img/tyko-logo.png"`
	Namespace       string `env:"CLUSTER_NAMESPACE" envDefault:"default"`
	ApiKey          string `env:"TYPESENSE_API_KEY,required"`
	Protocol        string `env:"TYPESENSE_PROTOCOL" envDefault:"http"`
	ApiPort         uint   `env:"TYPESENSE_API_PORT" envDefault:"8108"`
	PeeringPort     uint   `env:"TYPESENSE_PEERING_PORT" envDefault:"8107"`
	HealthCheckPort uint   `env:"HEALTHCHECK_PORT" envDefault:"8808"`
	NodesPath       string `env:"TYPESENSE_NODES" envDefault:"/usr/share/typesense/nodes"`
}
