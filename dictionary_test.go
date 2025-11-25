package stardict

import common "codeberg.org/ilius/go-dict-commons"

func init() {
	var _ common.Dictionary = &dictionaryImp{}
}
