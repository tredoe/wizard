package {{packageName}}

import (
	"log"
)

/*var App = Application {
	name:    "{{packageName}}",
	version: "{{version}}",
	summary: "{{summary}}",
	license: "{{license}}",
}*/


/* `env` should be 'development', 'test' or 'production'. */
func Setup(env string) {
	if env != "development" && env != "test" && env != "production" {
		log.Exitf("{{packageName}}.Setup\n  Invalid environment. Usage: 'development', 'test', or 'production'\n")
	}

	loadConfig(env)
	loadLocal()

	loadView()
	initRoute()
}

