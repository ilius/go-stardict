package stardict

import "log"

var ErrorHandler = func(err error) {
	log.Println(err)
}
