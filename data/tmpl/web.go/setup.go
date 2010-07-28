package {{pkg}}

import (
	"log"
)

var App = Application {
	name:        "{{pkg}}",
	version:     "",
	description: "",
	license:     "{{license}}",
}


/* `env` should be 'development', 'test' or 'production'. */
func Setup(env string) {
	if env != "development" && env != "test" && env != "production" {
		log.Exitf("{{pkg}}.Setup\n  Invalid environment. Usage: 'development', 'test', or 'production'\n")
	}

	loadConfig(env)
	loadLocal()

	loadView()
	initRoute()
}
