package {{application_name}}

import (
	"log"
)


/* `env` should be 'development', 'test' or 'production'. */
func Setup(env string) {
	if env != "development" && env != "test" && env != "production" {
		log.Exitf("{{application_name}}.Setup\n  Invalid environment. Usage: 'development', 'test', or 'production'\n")
	}

	loadConfig(env)
	loadLocal()

	loadView()
	initRoute()
}

