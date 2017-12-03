package factory

import (
	"log"
	"os"
)

// DebugMode is a flag controlling whether debug information is outputted to the os.Stdout
var DebugMode = false

var info = log.New(os.Stdout, "factory INFO ", log.Ldate|log.Ltime|log.Lshortfile)
