
# 🚨 REPOSITORY MOVED

This project has migrated to **Codeberg** for better alignment with open-source values.

## 👉 New Location: [codeberg.org/ilius/go-stardict](https://codeberg.org/ilius/go-stardict)

To download and install this package run:

`go get -u github.com/ilius/go-stardict`

Source docs: <http://godoc.org/github.com/ilius/go-stardict>

Disclaimer Sample code can be found in [`dict_test.go`](https://github.com/ilius/go-stardict/blob/master/dict_test.go).

## Project Overview

The project was started as an attempt to read stardict dictionaries in language learning webservice and grew into a tool supporting several dictionary formats.

Current limitations:

- Index file is loaded into memory for fast random access
- DictZip format is not supported, it is processed as a simple GZip format (means that no random blocks access is supported as in DictZip)
- There's no recovering from errors (means that dictionaries should be well formed)

Not tested but should be working in theory (I didn't find dictionaries with those properties in place):

- 64bit offsets
- multi typed dictionary fields

## Thanks

- [Format for StarDict dictionary files](http://www.stardict.org/StarDictFileFormat)
- [Dictionaries](http://download.huzheng.org/)
- [GO STARDICT](https://github.com/dyatlov/gostardict)
