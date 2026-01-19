package environment

import "database/sql"

var LoadSessionDelay = 1000
var CheckMailDelay = 1000
var MaxSessions = 100
var MaxInvalid = 1
var DB *sql.DB
