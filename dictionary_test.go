package stardict

import common "github.com/ilius/go-dict-commons"

func init() {
	var _ common.Dictionary = &dictionaryImp{}
}
