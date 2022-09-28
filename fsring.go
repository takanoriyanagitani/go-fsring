package fsring

type Read func(filename string) (data []byte, e error)

type List func(dirname string) (filenames []string, e error)
